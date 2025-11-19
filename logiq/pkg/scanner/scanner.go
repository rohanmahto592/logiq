package scanner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"logiq/pkg/config"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

type LogLine struct {
	FilePath  string `parquet:"name=file_path, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	LineNum   int32  `parquet:"name=line_num, type=INT32"`
	Content   string `parquet:"name=content, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	TimeStamp string `parquet:"name=timestamp, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

var stateFile = filepath.Join(".", "logiq_state.json")

type FileState struct {
	Offset   int64 `json:"offset"`
	LastLine int   `json:"last_line"`
}

func compilePatterns(patterns []string) []*regexp.Regexp {
	var regexes []*regexp.Regexp
	for _, p := range patterns {
		r, err := regexp.Compile(strings.TrimSpace(p))
		if err == nil {
			regexes = append(regexes, r)
		}
	}
	return regexes
}
func extractTimestamp(line string, timestampRegexes []*regexp.Regexp) string {
	for _, re := range timestampRegexes {
		if match := re.FindString(line); match != "" {
			return match
		}
	}
	return "" // fallback if no timestamp found
}

func loadOffsets() (map[string]FileState, error) {
	states := make(map[string]FileState)
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return states, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &states); err != nil {
		return nil, err
	}
	return states, nil
}
func saveOffsets(states map[string]FileState) error {
	data, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(stateFile, data, 0644)
}

func readFile(filePath string, includeRegexes, excludeRegexes []*regexp.Regexp, timestampRegexes []*regexp.Regexp, states map[string]FileState) ([]LogLine, FileState, error) {
	state := states[filePath]

	file, err := os.Open(filePath)
	if err != nil {
		return nil, state, err
	}
	defer file.Close()

	// Seek to last read position (for incremental scanning)
	_, err = file.Seek(state.Offset, io.SeekStart)
	if err != nil {
		return nil, state, err
	}

	reader := bufio.NewReaderSize(file, 1024*1024) // 1MB buffer
	var lines []LogLine
	lineNum := state.LastLine

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, state, err
		}

		// handle partial line if EOF without newline
		if len(line) > 0 {
			lineNum++
			text := strings.TrimRight(line, "\r\n")

			if shouldInclude(text, includeRegexes, excludeRegexes) {
				timestamp := extractTimestamp(text, timestampRegexes)
				lines = append(lines, LogLine{
					FilePath:  filePath,
					LineNum:   int32(lineNum),
					Content:   text,
					TimeStamp: timestamp,
				})
			}
		}

		if err == io.EOF {
			break
		}
	}

	// Save new offset + line number
	offset, _ := file.Seek(0, io.SeekCurrent)
	state.Offset = offset
	state.LastLine = lineNum

	return lines, state, nil
}

func shouldInclude(line string, includeRegexes, excludeRegexes []*regexp.Regexp) bool {
	for _, ex := range excludeRegexes {
		if ex.MatchString(line) {
			return false
		}
	}
	if len(includeRegexes) == 0 {
		return true
	}
	for _, in := range includeRegexes {
		if in.MatchString(line) {
			return true
		}
	}
	return false
}

func ScanLogs(cfg *config.Config) ([]LogLine, error) {
	includeRegexes := compilePatterns(cfg.IncludePatterns)
	excludeRegexes := compilePatterns(cfg.ExcludePatterns)
	timestampRegexes := compilePatterns(cfg.TimestampPatterns)

	states, _ := loadOffsets()
	results := make([]LogLine, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	fileChan := make(chan string, 10)
	workerCount := runtime.NumCPU() // tune based on CPU cores

	// Worker logic
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				lines, newState, err := readFile(filePath, includeRegexes, excludeRegexes, timestampRegexes, states)
				if err != nil {
					fmt.Printf("Warning: Failed to read %s: %v\n", filePath, err)
					continue
				}
				mu.Lock()
				results = append(results, lines...)
				states[filePath] = newState
				mu.Unlock()
			}
		}()
	}

	// Feed files to workers
	for _, path := range cfg.LogPaths {
		files, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("invalid log path pattern %s: %w", path, err)
		}
		for _, file := range files {
			fileChan <- file
		}
	}
	close(fileChan)

	wg.Wait()

	// ✅ Batch write offsets once after all reads complete
	if err := saveOffsets(states); err != nil {
		fmt.Printf("Warning: Failed to save offsets: %v\n", err)
	}

	return results, nil
}
