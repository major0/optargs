#!/bin/bash

# Performance Validation Script for OptArgs
# This script runs performance regression tests and validates results

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BASELINE_FILE="$PROJECT_ROOT/performance_baselines.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check if baseline file exists
check_baseline_file() {
    if [[ ! -f "$BASELINE_FILE" ]]; then
        print_status $YELLOW "No baseline file found at $BASELINE_FILE"
        print_status $BLUE "Establishing initial baselines..."
        return 1
    fi
    return 0
}

# Function to establish baselines
establish_baselines() {
    print_status $BLUE "Establishing performance baselines..."
    cd "$PROJECT_ROOT"
    
    # Run baseline establishment test
    if go test -v -run=TestPerformanceBaselines_Establish; then
        print_status $GREEN "✓ Performance baselines established successfully"
        return 0
    else
        print_status $RED "✗ Failed to establish performance baselines"
        return 1
    fi
}

# Function to run performance regression tests
run_regression_tests() {
    print_status $BLUE "Running performance regression tests..."
    cd "$PROJECT_ROOT"
    
    # Run regression tests
    if go test -v -run=TestPerformanceRegression; then
        print_status $GREEN "✓ All performance regression tests passed"
        return 0
    else
        print_status $RED "✗ Performance regression tests failed"
        return 1
    fi
}

# Function to run memory leak detection
run_memory_leak_tests() {
    print_status $BLUE "Running memory leak detection tests..."
    cd "$PROJECT_ROOT"
    
    if go test -v -run=TestMemoryLeakDetection; then
        print_status $GREEN "✓ No memory leaks detected"
        return 0
    else
        print_status $RED "✗ Memory leak tests failed"
        return 1
    fi
}

# Function to run iterator efficiency tests
run_iterator_tests() {
    print_status $BLUE "Running iterator efficiency tests..."
    cd "$PROJECT_ROOT"
    
    if go test -v -run=TestIteratorEfficiencyRegression; then
        print_status $GREEN "✓ Iterator efficiency tests passed"
        return 0
    else
        print_status $RED "✗ Iterator efficiency tests failed"
        return 1
    fi
}

# Function to run full benchmark suite
run_benchmarks() {
    print_status $BLUE "Running full benchmark suite..."
    cd "$PROJECT_ROOT"
    
    # Run benchmarks and save results
    local benchmark_output="benchmark_results_$(date +%Y%m%d_%H%M%S).txt"
    
    if go test -bench=. -benchmem -run=^$ > "$benchmark_output" 2>&1; then
        print_status $GREEN "✓ Benchmark suite completed successfully"
        print_status $BLUE "Results saved to: $benchmark_output"
        
        # Show summary of key benchmarks
        print_status $BLUE "Key benchmark results:"
        grep -E "BenchmarkGetOpt|BenchmarkGetOptLong|BenchmarkMemoryAllocation" "$benchmark_output" | head -10
        
        return 0
    else
        print_status $RED "✗ Benchmark suite failed"
        cat "$benchmark_output"
        return 1
    fi
}

# Function to generate performance report
generate_report() {
    print_status $BLUE "Generating performance report..."
    cd "$PROJECT_ROOT"
    
    local report_file="performance_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$report_file" << EOF
# OptArgs Performance Report

Generated: $(date)
Go Version: $(go version)
System: $(uname -a)

## Performance Baselines

EOF
    
    if [[ -f "$BASELINE_FILE" ]]; then
        echo "Baseline file exists with $(jq '.baselines | length' "$BASELINE_FILE" 2>/dev/null || echo "unknown") test cases" >> "$report_file"
        echo "" >> "$report_file"
        echo "### Latest Baselines" >> "$report_file"
        echo "" >> "$report_file"
        
        # Extract key metrics if jq is available
        if command -v jq >/dev/null 2>&1; then
            jq -r '.baselines[] | "- \(.test_name): \(.ns_per_op) ns/op, \(.allocs_per_op) allocs/op, \(.bytes_per_op) bytes/op"' "$BASELINE_FILE" >> "$report_file"
        else
            echo "jq not available - raw baseline data:" >> "$report_file"
            cat "$BASELINE_FILE" >> "$report_file"
        fi
    else
        echo "No baseline file found" >> "$report_file"
    fi
    
    echo "" >> "$report_file"
    echo "## Test Results" >> "$report_file"
    echo "" >> "$report_file"
    
    # Add test results if available
    if [[ -f "benchmark_results_"*".txt" ]]; then
        local latest_results=$(ls -t benchmark_results_*.txt | head -1)
        echo "### Latest Benchmark Results" >> "$report_file"
        echo "" >> "$report_file"
        echo '```' >> "$report_file"
        cat "$latest_results" >> "$report_file"
        echo '```' >> "$report_file"
    fi
    
    print_status $GREEN "✓ Performance report generated: $report_file"
}

# Function to validate performance thresholds
validate_thresholds() {
    print_status $BLUE "Validating performance thresholds..."
    
    # This would typically compare against CI/CD performance requirements
    # For now, we'll just run the regression tests which have built-in thresholds
    if run_regression_tests; then
        print_status $GREEN "✓ All performance thresholds met"
        return 0
    else
        print_status $RED "✗ Performance thresholds exceeded"
        return 1
    fi
}

# Main execution
main() {
    print_status $BLUE "OptArgs Performance Validation"
    print_status $BLUE "=============================="
    
    local command=${1:-"all"}
    local exit_code=0
    
    case "$command" in
        "baseline"|"baselines")
            establish_baselines || exit_code=1
            ;;
        "regression")
            if ! check_baseline_file; then
                establish_baselines || exit_code=1
            fi
            run_regression_tests || exit_code=1
            ;;
        "memory")
            run_memory_leak_tests || exit_code=1
            ;;
        "iterator")
            run_iterator_tests || exit_code=1
            ;;
        "benchmark"|"benchmarks")
            run_benchmarks || exit_code=1
            ;;
        "report")
            generate_report || exit_code=1
            ;;
        "validate")
            validate_thresholds || exit_code=1
            ;;
        "all")
            # Run complete validation suite
            if ! check_baseline_file; then
                establish_baselines || exit_code=1
            fi
            
            if [[ $exit_code -eq 0 ]]; then
                run_regression_tests || exit_code=1
            fi
            
            if [[ $exit_code -eq 0 ]]; then
                run_memory_leak_tests || exit_code=1
            fi
            
            if [[ $exit_code -eq 0 ]]; then
                run_iterator_tests || exit_code=1
            fi
            
            if [[ $exit_code -eq 0 ]]; then
                run_benchmarks || exit_code=1
            fi
            
            generate_report
            ;;
        "help"|"-h"|"--help")
            cat << EOF
Usage: $0 [command]

Commands:
  baseline    Establish performance baselines
  regression  Run performance regression tests
  memory      Run memory leak detection tests
  iterator    Run iterator efficiency tests
  benchmark   Run full benchmark suite
  report      Generate performance report
  validate    Validate performance thresholds
  all         Run complete validation suite (default)
  help        Show this help message

Examples:
  $0                    # Run complete validation suite
  $0 baseline          # Establish new baselines
  $0 regression        # Check for performance regressions
  $0 benchmark         # Run benchmarks only
EOF
            ;;
        *)
            print_status $RED "Unknown command: $command"
            print_status $BLUE "Use '$0 help' for usage information"
            exit_code=1
            ;;
    esac
    
    if [[ $exit_code -eq 0 ]]; then
        print_status $GREEN "✓ Performance validation completed successfully"
    else
        print_status $RED "✗ Performance validation failed"
    fi
    
    exit $exit_code
}

# Run main function with all arguments
main "$@"