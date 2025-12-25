package main

import (
	"context"
	"os"
	"testing"

	"google.golang.org/adk/tool"
)

// Mock tool.Context for testing purposes
type mockToolContext struct{}

func (m *mockToolContext) Context() context.Context {
	return context.Background()
}

func TestWriteFileTool(t *testing.T) {
	ctx := &mockToolContext{}
	testFileName := "test_write.txt"
	defer os.Remove(testFileName) // Clean up after test

	// Test writing to a new file
	content1 := "Hello, world!"
	input1 := WriteInput{FileName: testFileName, Content: content1}
	result1, err1 := func(ctx tool.Context, input WriteInput) (string, error) {
		file, err := os.OpenFile(input.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return "Failed to open/create file " + input.FileName, err
		}
		defer file.Close()
		if _, err := file.WriteString(input.Content); err != nil {
			return "Failed to write the content to file " + input.FileName, err
		}
		return "Successfully written the content to file " + input.Content, nil
	}(ctx, input1)

	if err1 != nil {
		t.Fatalf("write_file failed unexpectedly: %v", err1)
	}
	expectedResult1 := "Successfully written the content to file " + content1
	if result1 != expectedResult1 {
		t.Errorf("Expected '%s', got '%s'", expectedResult1, result1)
	}

	readContent1, err := os.ReadFile(testFileName)
	if err != nil {
		t.Fatalf("Failed to read file after write: %v", err)
	}
	if string(readContent1) != content1 {
		t.Errorf("File content mismatch. Expected '%s', got '%s'", content1, string(readContent1))
	}

	// Test appending to an existing file
	content2 := " Appended content."
	input2 := WriteInput{FileName: testFileName, Content: content2}
	result2, err2 := func(ctx tool.Context, input WriteInput) (string, error) {
		file, err := os.OpenFile(input.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return "Failed to open/create file " + input.FileName, err
		}
		defer file.Close()
		if _, err := file.WriteString(input.Content); err != nil {
			return "Failed to write the content to file " + input.FileName, err
		}
		return "Successfully written the content to file " + input.Content, nil
	}(ctx, input2)

	if err2 != nil {
		t.Fatalf("write_file (append) failed unexpectedly: %v", err2)
	}
	expectedResult2 := "Successfully written the content to file " + content2
	if result2 != expectedResult2 {
		t.Errorf("Expected '%s', got '%s'", expectedResult2, result2)
	}

	readContent2, err := os.ReadFile(testFileName)
	if err != nil {
		t.Fatalf("Failed to read file after append: %v", err)
	}
	expectedFullContent := content1 + content2
	if string(readContent2) != expectedFullContent {
		t.Errorf("File content mismatch after append. Expected '%s', got '%s'", expectedFullContent, string(readContent2))
	}
}

func TestReadFileTool(t *testing.T) {
	ctx := &mockToolContext{}
	testFileName := "test_read.txt"
	defer os.Remove(testFileName) // Clean up after test

	// Test reading from an existing file
	content := "This is a test file for reading."
	err := os.WriteFile(testFileName, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	input1 := ReadInput{FileName: testFileName}
	result1, err1 := func(ctx tool.Context, input ReadInput) (string, error) {
		var content_bytes []byte
		if content_bytes, err = os.ReadFile(input.FileName); err != nil {
			return "Failed to read the content of the file " + input.FileName, err
		}
		return string(content_bytes), nil
	}(ctx, input1)

	if err1 != nil {
		t.Fatalf("read_file failed unexpectedly: %v", err1)
	}
	if result1 != content {
		t.Errorf("Expected '%s', got '%s'", content, result1)
	}

	// Test reading from a non-existent file
	nonExistentFileName := "non_existent_file.txt"
	input2 := ReadInput{FileName: nonExistentFileName}
	result2, err2 := func(ctx tool.Context, input ReadInput) (string, error) {
		var content_bytes []byte
		if content_bytes, err = os.ReadFile(input.FileName); err != nil {
			return "Failed to read the content of the file " + input.FileName, err
		}
		return string(content_bytes), nil
	}(ctx, input2)

	if err2 == nil {
		t.Fatalf("read_file did not return an error for non-existent file")
	}
	expectedErrorSubstring := "Failed to read the content of the file " + nonExistentFileName
	if result2 != expectedErrorSubstring {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrorSubstring, result2)
	}

	// Test reading from an empty file
	emptyFileName := "test_empty.txt"
	err = os.WriteFile(emptyFileName, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty test file: %v", err)
	}
	defer os.Remove(emptyFileName)

	input3 := ReadInput{FileName: emptyFileName}
	result3, err3 := func(ctx tool.Context, input ReadInput) (string, error) {
		var content_bytes []byte
		if content_bytes, err = os.ReadFile(input.FileName); err != nil {
			return "Failed to read the content of the file " + input.FileName, err
		}
		return string(content_bytes), nil
	}(ctx, input3)

	if err3 != nil {
		t.Fatalf("read_file failed unexpectedly for empty file: %v", err3)
	}
	if result3 != "" {
		t.Errorf("Expected empty string, got '%s'", result3)
	}
}
