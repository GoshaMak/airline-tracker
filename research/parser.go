package main

import (
	"fmt"
	"regexp"
	"strconv"
)

func ExtractExplainTimes(s string) (planningTime float64, executionTime float64, err error) {
	planningRe := regexp.MustCompile(`Planning Time:\s*([0-9]+(?:\.[0-9]+)?)\s*ms`)
	executionRe := regexp.MustCompile(`Execution Time:\s*([0-9]+(?:\.[0-9]+)?)\s*ms`)

	planningMatch := planningRe.FindStringSubmatch(s)
	if len(planningMatch) < 2 {
		return 0, 0, fmt.Errorf("planning time not found")
	}

	executionMatch := executionRe.FindStringSubmatch(s)
	if len(executionMatch) < 2 {
		return 0, 0, fmt.Errorf("execution time not found")
	}

	planningTime, err = strconv.ParseFloat(planningMatch[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse planning time: %w", err)
	}

	executionTime, err = strconv.ParseFloat(executionMatch[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse execution time: %w", err)
	}

	return planningTime, executionTime, nil
}

func ExtractPlanningTime(s string) (planningTime float64, err error) {
	planningRe := regexp.MustCompile(`Planning Time:\s*([0-9]+(?:\.[0-9]+)?)\s*ms`)

	planningMatch := planningRe.FindStringSubmatch(s)
	if len(planningMatch) < 2 {
		return 0, fmt.Errorf("planning time not found")
	}

	planningTime, err = strconv.ParseFloat(planningMatch[1], 64)
	if err != nil {
		return 0, fmt.Errorf("parse planning time: %w", err)
	}

	return planningTime, nil
}
func ExtractExecutionTime(s string) (executionTime float64, err error) {
	executionRe := regexp.MustCompile(`Execution Time:\s*([0-9]+(?:\.[0-9]+)?)\s*ms`)

	executionMatch := executionRe.FindStringSubmatch(s)
	if len(executionMatch) < 2 {
		return 0, fmt.Errorf("execution time not found")
	}

	executionTime, err = strconv.ParseFloat(executionMatch[1], 64)
	if err != nil {
		return 0, fmt.Errorf("parse execution time: %w", err)
	}

	return executionTime, nil
}
