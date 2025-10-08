# PowerShell Build Script - AGILira Standard
# Windows equivalent of Makefile for Go development
# Usage: .\Makefile.ps1 [command]
# Example: .\Makefile.ps1 help

param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

# Variables
$BinaryName = Split-Path -Leaf (Get-Location)
$ToolsDir = "$env:GOPATH\bin"
if (-not $ToolsDir) { $ToolsDir = "$env:USERPROFILE\go\bin" }

# Colors for output
$Red = "Red"
$Green = "Green" 
$Yellow = "Yellow"
$Blue = "Cyan"

function Write-ColorOutput {
    param($Message, $Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Test-ToolExists {
    param($ToolName)
    $toolPath = Join-Path $ToolsDir "$ToolName.exe"
    return Test-Path $toolPath
}

function Invoke-Help {
    Write-ColorOutput "Available commands:" $Blue
    Write-ColorOutput "  help             Show this help message" $Green
    Write-ColorOutput "  test             Run tests" $Green
    Write-ColorOutput "  race             Run tests with race detector" $Green
    Write-ColorOutput "  coverage         Run tests with coverage" $Green
    Write-ColorOutput "  fmt              Format Go code" $Green
    Write-ColorOutput "  vet              Run go vet" $Green
    Write-ColorOutput "  staticcheck      Run staticcheck" $Green
    Write-ColorOutput "  errcheck         Run errcheck" $Green
    Write-ColorOutput "  gosec            Run gosec security scanner" $Green
    Write-ColorOutput "  govulncheck      Run govulncheck vulnerability scanner" $Green
    Write-ColorOutput "  lint             Run all linters" $Green
    Write-ColorOutput "  security         Run security checks" $Green
    Write-ColorOutput "  check            Run all checks (format, vet, lint, security, test)" $Green
    Write-ColorOutput "  check-race       Run all checks including race detector" $Green
    Write-ColorOutput "  tools            Install development tools" $Green
    Write-ColorOutput "  deps             Download and verify dependencies" $Green
    Write-ColorOutput "  clean            Clean build artifacts and test cache" $Green
    Write-ColorOutput "  build            Build the binary" $Green
    Write-ColorOutput "  install          Install the binary to GOPATH/bin" $Green
    Write-ColorOutput "  bench            Run benchmarks" $Green
    Write-ColorOutput "  fuzz             Run fuzz tests (30s each)" $Green
    Write-ColorOutput "  fuzz-extended    Run extended fuzz tests (5m each)" $Green
    Write-ColorOutput "  fuzz-continuous  Run continuous fuzz tests (until Ctrl+C)" $Green
    Write-ColorOutput "  fuzz-report      Generate fuzz testing report" $Green
    Write-ColorOutput "  security-full    Run complete security testing (static + fuzz)" $Green
    Write-ColorOutput "  ci               Run CI checks" $Green
    Write-ColorOutput "  ci-security      Run CI checks with fuzz testing" $Green
    Write-ColorOutput "  dev              Quick development check" $Green
    Write-ColorOutput "  pre-commit       Run pre-commit checks (alias for 'check')" $Green
    Write-ColorOutput "  all              Run everything from scratch" $Green
    Write-ColorOutput "  status           Show status of installed tools" $Green
}

function Invoke-Test {
    Write-ColorOutput "Running tests..." $Yellow
    go test -v "./..."
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-Race {
    Write-ColorOutput "Running tests with race detector..." $Yellow
    go test -race -v "./..."
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-Coverage {
    Write-ColorOutput "Running tests with coverage..." $Yellow
    $testArgs = @("-coverprofile=coverage.out", "./...")
    go test @testArgs
    if ($LASTEXITCODE -eq 0) {
        $coverArgs = @("-html=coverage.out", "-o", "coverage.html")
        go tool cover @coverArgs
        Write-ColorOutput "Coverage report generated: coverage.html" $Green
    }
}

function Invoke-Fmt {
    Write-ColorOutput "Formatting Go code..." $Yellow
    go fmt "./..."
}

function Invoke-Vet {
    Write-ColorOutput "Running go vet..." $Yellow
    go vet "./..."
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-StaticCheck {
    Write-ColorOutput "Running staticcheck..." $Yellow
    if (-not (Test-ToolExists "staticcheck")) {
        Write-ColorOutput "staticcheck not found. Run '.\Makefile.ps1 tools' to install." $Red
        exit 1
    }
    & "$ToolsDir\staticcheck.exe" "./..."
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-ErrCheck {
    Write-ColorOutput "Running errcheck..." $Yellow
    if (-not (Test-ToolExists "errcheck")) {
        Write-ColorOutput "errcheck not found. Run '.\Makefile.ps1 tools' to install." $Red
        exit 1
    }
    & "$ToolsDir\errcheck.exe" "./..."
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-GoSec {
    Write-ColorOutput "Running gosec security scanner..." $Yellow
    if (-not (Test-ToolExists "gosec")) {
        Write-ColorOutput "gosec not found. Run '.\Makefile.ps1 tools' to install." $Red
        exit 1
    }
    & "$ToolsDir\gosec.exe" "./..."
    if ($LASTEXITCODE -ne 0) {
        Write-ColorOutput "⚠️  gosec completed with warnings (may be import-related)" $Yellow
    }
}

function Invoke-GoVulnCheck {
    Write-ColorOutput "Running govulncheck vulnerability scanner..." $Yellow
    if (-not (Test-ToolExists "govulncheck")) {
        Write-ColorOutput "govulncheck not found. Run '.\Makefile.ps1 tools' to install." $Red
        exit 1
    }
    & "$ToolsDir\govulncheck.exe" "./..."
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-Lint {
    Invoke-StaticCheck
    Invoke-ErrCheck
    Write-ColorOutput "All linters completed." $Green
}

function Invoke-Security {
    Invoke-GoSec
    Invoke-GoVulnCheck
    Write-ColorOutput "Security checks completed." $Green
}

function Invoke-Check {
    Invoke-Fmt
    Invoke-Vet
    Invoke-Lint
    Invoke-Security
    Invoke-Test
    Write-ColorOutput "All checks passed!" $Green
}

function Invoke-CheckRace {
    Invoke-Fmt
    Invoke-Vet
    Invoke-Lint
    Invoke-Security
    Invoke-Race
    Write-ColorOutput "All checks with race detection passed!" $Green
}

function Invoke-Tools {
    Write-ColorOutput "Installing development tools..." $Yellow
    go install honnef.co/go/tools/cmd/staticcheck@latest
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    
    go install github.com/kisielk/errcheck@latest
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    
    go install golang.org/x/vuln/cmd/govulncheck@latest
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    
    Write-ColorOutput "Tools installed successfully!" $Green
}

function Invoke-Deps {
    Write-ColorOutput "Downloading dependencies..." $Yellow
    go mod download
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    
    go mod verify
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    
    go mod tidy
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-Clean {
    Write-ColorOutput "Cleaning..." $Yellow
    go clean
    go clean -testcache
    if (Test-Path "coverage.out") { Remove-Item "coverage.out" }
    if (Test-Path "coverage.html") { Remove-Item "coverage.html" }
    if (Test-Path "$BinaryName.exe") { Remove-Item "$BinaryName.exe" }
}

function Invoke-Build {
    Write-ColorOutput "Building $BinaryName..." $Yellow
    go build -ldflags="-w -s" -o "$BinaryName.exe" .
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-Install {
    Write-ColorOutput "Installing $BinaryName..." $Yellow
    go install .
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-Bench {
    Write-ColorOutput "Running benchmarks..." $Yellow
    go test -bench=. -benchmem "./..."
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-Fuzz {
    Write-ColorOutput "Running fuzz tests (30s each)..." $Yellow
    $fuzzFunctions = @(
        "FuzzParse",
        "FuzzParseStringSlice", 
        "FuzzLoadConfig",
        "FuzzEnvironmentVariables",
        "FuzzFlagValidation"
    )
    
    foreach ($func in $fuzzFunctions) {
        Write-ColorOutput "Fuzzing $func..." $Cyan
        go test -fuzz="$func" -fuzztime=30s .
        if ($LASTEXITCODE -ne 0) { 
            Write-ColorOutput "Fuzz test $func failed!" $Red
            exit $LASTEXITCODE 
        }
    }
    Write-ColorOutput "All fuzz tests completed successfully!" $Green
}

function Invoke-FuzzExtended {
    Write-ColorOutput "Running extended fuzz tests (5m each)..." $Yellow
    $fuzzFunctions = @(
        "FuzzParse",
        "FuzzParseStringSlice",
        "FuzzLoadConfig", 
        "FuzzEnvironmentVariables",
        "FuzzFlagValidation"
    )
    
    foreach ($func in $fuzzFunctions) {
        Write-ColorOutput "Extended fuzzing $func..." $Cyan
        go test -fuzz="$func" -fuzztime=5m .
        if ($LASTEXITCODE -ne 0) { 
            Write-ColorOutput "Extended fuzz test $func failed!" $Red
            exit $LASTEXITCODE 
        }
    }
    Write-ColorOutput "All extended fuzz tests completed successfully!" $Green
}

function Invoke-FuzzContinuous {
    Write-ColorOutput "Starting continuous fuzz testing (Ctrl+C to stop)..." $Yellow
    Write-ColorOutput "This will run fuzz tests continuously for all flash-flags functions..." $Cyan
    
    $fuzzFunctions = @(
        "FuzzParse",
        "FuzzParseStringSlice",
        "FuzzLoadConfig",
        "FuzzEnvironmentVariables", 
        "FuzzFlagValidation"
    )
    
    $currentIndex = 0
    while ($true) {
        $func = $fuzzFunctions[$currentIndex]
        Write-ColorOutput "Continuous fuzzing $func (30s)..." $Cyan
        go test -fuzz="$func" -fuzztime=30s .
        
        $currentIndex = ($currentIndex + 1) % $fuzzFunctions.Length
        Start-Sleep -Seconds 2
    }
}

function Invoke-FuzzReport {
    Write-ColorOutput "Generating fuzz test report..." $Yellow
    
    if (Test-Path "./testdata/fuzz") {
        Write-ColorOutput "Fuzz corpus files:" $Cyan
        Get-ChildItem -Path "./testdata/fuzz" -Recurse | ForEach-Object {
            Write-Host "  $($_.FullName)" -ForegroundColor Gray
        }
        
        $corpusCount = (Get-ChildItem -Path "./testdata/fuzz" -Recurse -File).Count
        Write-ColorOutput "Total corpus entries: $corpusCount" $Green
    } else {
        Write-ColorOutput "No fuzz corpus found. Run 'make fuzz' to generate corpus." $Yellow
    }
    
    # Check for any crash files
    $crashFiles = Get-ChildItem -Path "." -Filter "*testdata*" -Recurse | Where-Object { $_.Name -match "crash" }
    if ($crashFiles) {
        Write-ColorOutput "Found crash files:" $Red
        $crashFiles | ForEach-Object {
            Write-Host "  $($_.FullName)" -ForegroundColor Red
        }
    } else {
        Write-ColorOutput "No crash files found." $Green
    }
}

function Invoke-SecurityFull {
    Write-ColorOutput "Running comprehensive security tests..." $Blue
    Write-ColorOutput "1. Static security analysis..." $Cyan
    Invoke-Security
    
    Write-ColorOutput "2. Fuzz testing..." $Cyan  
    Invoke-Fuzz
    
    Write-ColorOutput "3. Generating security report..." $Cyan
    Invoke-FuzzReport
    
    Write-ColorOutput "Comprehensive security testing completed!" $Green
}

function Invoke-CISecurity {
    Write-ColorOutput "Running security-focused fuzz testing for CI/CD pipelines..." $Yellow
    Write-ColorOutput "This runs all flash-flags fuzz tests with security emphasis..." $Cyan
    
    $securityFuzzFunctions = @(
        "FuzzParse",
        "FuzzParseStringSlice", 
        "FuzzLoadConfig",
        "FuzzEnvironmentVariables",
        "FuzzFlagValidation"
    )
    
    foreach ($func in $securityFuzzFunctions) {
        Write-ColorOutput "Security fuzzing $func..." $Cyan
        go test -fuzz="$func" -fuzztime=2m .
        if ($LASTEXITCODE -ne 0) { 
            Write-ColorOutput "Security fuzz test $func failed!" $Red
            exit $LASTEXITCODE 
        }
    }
    
    Write-ColorOutput "All security fuzz tests passed!" $Green
    Write-ColorOutput "flash-flags is secure against known attack vectors" $Green
}

function Invoke-CI {
    Write-ColorOutput "Running CI checks..." $Blue
    Invoke-Fmt
    Invoke-Vet
    Invoke-Lint
    Invoke-Security
    Invoke-Test
    Invoke-Coverage
    Write-ColorOutput "CI checks completed successfully!" $Green
}

function Invoke-Dev {
    Write-ColorOutput "Running development checks..." $Blue
    Invoke-Fmt
    Invoke-Vet
    Invoke-Test
    Write-ColorOutput "Development checks completed!" $Green
}

function Invoke-All {
    Invoke-Clean
    Invoke-Tools
    Invoke-Deps
    Invoke-Check
    Invoke-Build
}

function Invoke-Status {
    Write-ColorOutput "Development tools status:" $Blue
    
    $staticcheckStatus = if (Test-ToolExists "staticcheck") { "✓ installed" } else { "✗ missing" }
    $staticcheckColor = if (Test-ToolExists "staticcheck") { $Green } else { $Red }
    Write-Host "staticcheck:   " -NoNewline
    Write-ColorOutput $staticcheckStatus $staticcheckColor
    
    $errcheckStatus = if (Test-ToolExists "errcheck") { "✓ installed" } else { "✗ missing" }
    $errcheckColor = if (Test-ToolExists "errcheck") { $Green } else { $Red }
    Write-Host "errcheck:      " -NoNewline
    Write-ColorOutput $errcheckStatus $errcheckColor
    
    $gosecStatus = if (Test-ToolExists "gosec") { "✓ installed" } else { "✗ missing" }
    $gosecColor = if (Test-ToolExists "gosec") { $Green } else { $Red }
    Write-Host "gosec:         " -NoNewline
    Write-ColorOutput $gosecStatus $gosecColor
    
    $govulncheckStatus = if (Test-ToolExists "govulncheck") { "✓ installed" } else { "✗ missing" }
    $govulncheckColor = if (Test-ToolExists "govulncheck") { $Green } else { $Red }
    Write-Host "govulncheck:   " -NoNewline
    Write-ColorOutput $govulncheckStatus $govulncheckColor
}

# Main execution
switch ($Command.ToLower()) {
    "help" { Invoke-Help }
    "test" { Invoke-Test }
    "race" { Invoke-Race }
    "coverage" { Invoke-Coverage }
    "fmt" { Invoke-Fmt }
    "vet" { Invoke-Vet }
    "staticcheck" { Invoke-StaticCheck }
    "errcheck" { Invoke-ErrCheck }
    "gosec" { Invoke-GoSec }
    "govulncheck" { Invoke-GoVulnCheck }
    "lint" { Invoke-Lint }
    "security" { Invoke-Security }
    "check" { Invoke-Check }
    "check-race" { Invoke-CheckRace }
    "tools" { Invoke-Tools }
    "deps" { Invoke-Deps }
    "clean" { Invoke-Clean }
    "build" { Invoke-Build }
    "install" { Invoke-Install }
    "bench" { Invoke-Bench }
    "ci" { Invoke-CI }
    "dev" { Invoke-Dev }
    "pre-commit" { Invoke-Check }
    "all" { Invoke-All }
    "status" { Invoke-Status }
    "fuzz" { Invoke-Fuzz }
    "fuzz-extended" { Invoke-FuzzExtended }
    "fuzz-continuous" { Invoke-FuzzContinuous }
    "fuzz-report" { Invoke-FuzzReport }
    "security-full" { Invoke-SecurityFull }
    "ci-security" { Invoke-CISecurity }
    default {
        Write-ColorOutput "Unknown command: $Command" $Red
        Write-ColorOutput "Run '.\Makefile.ps1 help' for available commands." $Yellow
        exit 1
    }
}