package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func ParseCommandParam(arg string, param interface{}) error {
	if arg == "" {
		return errors.New("empty command")
	}
	switch param := param.(type) {
	case *ReadFileParam:
		param.FilePath = ExtractTag(arg, "filepath")
		if param.FilePath == "" {
			return errors.New("filepath parameter not found, please re-output the command in the format specified")
		}
		param.StartLine = 0
		param.EndLine = 0
		startarg := ExtractTag(arg, "startline")
		endarg := ExtractTag(arg, "endline")
		if startarg != "" && endarg != "" {
			startliene, err := strconv.Atoi(startarg)
			if err != nil {
				return err
			}
			endline, err := strconv.Atoi(endarg)
			if err != nil {
				return err
			}
			param.StartLine = startliene
			param.EndLine = endline
		}
	case *ListDirParam:
		param.FilePath = ExtractTag(arg, "filepath")
		if param.FilePath == "" {
			return errors.New("filepath parameter not found, please re-output the command in the format specified")
		}
		param.Depth = 3
		param.Exts = ExtractTag(arg, "exts")
		depthArg := ExtractTag(arg, "depth")
		if depthArg != "" {
			depth, err := strconv.Atoi(depthArg)
			if err != nil {
				return err
			}
			param.Depth = depth
		}
	case *GrepParam:
		param.FilePath = ExtractTag(arg, "filepath")
		if param.FilePath == "" {
			return errors.New("filepath parameter not found, please re-output the command in the format specified")
		}
		param.Pattern = ExtractTag(arg, "regex")
		param.Pattern = strings.ReplaceAll(param.Pattern, "&lt;", "<")
		param.Pattern = strings.ReplaceAll(param.Pattern, "&gt;", ">")
		param.Context = 3
		contextArg := ExtractTag(arg, "contextline")
		if contextArg != "" {
			context2, err := strconv.Atoi(contextArg)
			if err != nil {
				return err
			}
			param.Context = context2
		}
	}
	return nil
}

func ExtractTag(text, tag string) string {
	startText := fmt.Sprintf("<%s>", tag)
	endText := fmt.Sprintf("</%s>", tag)
	startIndex := strings.Index(text, startText)
	if startIndex == -1 {
		return ""
	}
	tmp := text[startIndex+len(startText):]
	if !strings.Contains(tmp, endText) {
		return ""
	}
	endIndex := strings.Index(tmp, endText) + startIndex + len(startText)
	if startIndex == -1 || endIndex == -1 || endIndex <= startIndex {
		return ""
	}
	return strings.TrimSpace(text[startIndex+len(startText) : endIndex])
}
