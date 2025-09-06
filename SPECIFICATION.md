# URLPicker Specification

## Purpose

**urlpicker** is a command-line utility that extracts syntactically valid absolute URLs from arbitrary plain text and transforms Git remote specifications into human-navigable HTTPS repository URLs.

### Primary Goals
- Extract all RFC 3986 compliant absolute URLs from stdin
- Transform Git remote forms to canonical HTTPS repository URLs
- Output normalized URLs to stdout, one per line
- Provide deterministic, high-performance processing of terminal buffers and logs

### Interface Contract
- **Input**: Complete stdin stream (UTF-8 text)
- **Output**: Newline-delimited normalized URLs to stdout
- **Interaction**: Zero - no flags, prompts, or configuration
- **Exit Code**: 0 on success, non-zero only on fatal I/O errors

## Scope & References

### In Scope
- **RFC 3986** absolute URIs: `https://tools.ietf.org/rfc/rfc3986.txt`
- **Git URL formats** per official documentation: `https://git-scm.com/docs/git-clone#_git_urls`
- Git remote transformation to canonical repository web URLs
- Deduplication of URLs

### Out of Scope
- Content retractions (e.g. credential removal)
- Relative URLs or schemeless hostnames
- Interactive usage, configuration files, or command-line flags
- Network validation or content fetching
- Metadata output (JSON, confidence scores, offsets)
- Branch/path inference beyond repository root

### Official References
1. **RFC 3986 - Uniform Resource Identifier (URI): Generic Syntax**
   - https://tools.ietf.org/rfc/rfc3986.txt
   - Sections 3.1 (scheme), 3.2 (authority), 3.3 (path), 3.4 (query), 3.5 (fragment)

2. **Git Clone URL Documentation**
   - https://git-scm.com/docs/git-clone#_git_urls
   - Covers SSH, SCP-like, git://, and HTTP(S) remote formats

## Terminology

| Term | Definition |
|------|------------|
| **Absolute URL** | RFC 3986 compliant URI with required scheme component |
| **Git Remote** | Any syntax accepted by `git clone` command |
| **SCP-like Remote** | `[user@]host.xz:path/to/repo(.git)?` format (no `://`) |
| **Canonical Repo URL** | `https://host/owner/repo` (no trailing slash, no `.git`) |
| **Wrapper Punctuation** | Surrounding punctuation not part of URL: `.,;:!?)]}'"` |

## Functional Requirements

### Input Processing (FR-INPUT)

**FR-INPUT-1**: Read complete stdin as UTF-8 bytes
- Invalid UTF-8 sequences may be skipped or replaced
- No encoding errors abort processing

**FR-INPUT-2**: Process input synchonicity
- Streaming or incremental output is allowed, but not required
- Detected URLs output any time after detection
- No duplicate output independent of sync/async decision

**FR-INPUT-3**: Handle arbitrary text content
- Terminal buffers with ANSI escape codes
- Mixed punctuation and control characters
- URLs are not split across multiple lines
- No assumptions about structure or format

### URL Detection (FR-DETECT)

**FR-DETECT-1**: Identify RFC 3986 absolute URIs  
- Accept only HTTP and HTTPS schemes (web-focused extraction)
- Authority must contain valid host for hierarchical URIs
- Other schemes (ftp, file, mailto, etc.) are excluded by design

**FR-DETECT-2**: Detect Git remote forms
```
# SCP-like
git@github.com:owner/repo.git

# SSH explicit scheme  
ssh://user@host:port/path/to/repo.git

# Git protocol
git://host/path/to/repo.git

# HTTPS/HTTP
https://github.com/owner/repo.git
```

**FR-DETECT-3**: Candidate boundary rules
- Terminate at first unescaped whitespace or control character (≤ 0x1F, 0x7F)
- Exclude incomplete percent-escape sequences (`%` or `%[0-9A-Fa-f]` at end)
- Apply wrapper punctuation trimming

### Normalization (FR-NORM)

**FR-NORM-1**: Host normalization
- Convert to lowercase per RFC 3986
- Apply IDNA punycode for internationalized domain names

**FR-NORM-2**: Wrapper punctuation trimming
- Remove trailing: `.,;:!?'"""')]}>»›`
- Preserve balanced internal parentheses/brackets
- Iterative removal while maintaining valid scheme

**FR-NORM-3**: Percent-encoding validation
- Accept only complete `%HH` sequences
- Reject candidates with trailing incomplete escapes
- Do not attempt repair of malformed sequences

### Git Remote Transformation (FR-GIT)

**FR-GIT-1**: SCP-like transformation
```
Input:  git@github.com:owner/repo.git
Output: https://github.com/owner/repo
```

**FR-GIT-2**: SSH scheme transformation  
```
Input:  ssh://user@host:2222/group/repo.git
Output: https://host/group/repo
```

**FR-GIT-3**: Git protocol transformation
```
Input:  git://host/path/to/repo.git  
Output: https://host/path/to/repo
```

**FR-GIT-4**: HTTPS remote normalization
```
Input:  https://gitlab.com/group/project.git
Output: https://gitlab.com/group/project
```

**FR-GIT-5**: Transformation rules
- Always output `https://` scheme
- Remove user component and port
- Strip trailing `.git` (case-insensitive)
- Remove trailing `.git/` directory references
- Preserve nested group/organization paths
- No host whitelist (transform any valid remote)

### Output (FR-OUTPUT)

**FR-OUTPUT-1**: Format specification
- One normalized URL per line
- UTF-8 encoding
- Unix line endings (`\n`)
- No trailing blank line

**FR-OUTPUT-2**: Deduplication
- Emit each unique normalized URL exactly once
- Preserve first-occurrence order
- Apply after all normalization and transformation

**FR-OUTPUT-3**: Deterministic ordering
- Same input produces identical output across runs
- Stable sort behavior independent of locale

## Non-Functional Requirements

### Performance (NFR-PERF)
- **Throughput**: ≥ 40 MB/s on commodity laptop for mixed text
- **Latency**: 10 MB buffer processed in < 250ms
- **Memory**: ≤ input_size + O(unique_urls × average_length)
- **Algorithm**: No quadratic time complexity on adversarial input

### Reliability (NFR-REL)
- **Robustness**: No panics on arbitrary byte input
- **Security**: Never resolve DNS or open network connections
- **Privacy**: Never echo or log credentials
- **Error Handling**: Graceful degradation on partial I/O errors

### Implementation (NFR-IMPL)
- **Simplicity**: Core implementation < 400 LOC
- **Dependencies**: Standard library only (Go)
- **Portability**: Single static binary
- **Determinism**: Identical output across platforms and runs

## Algorithm Specification

### Extraction Pipeline
1. **Scan Phase**: Identify candidate substrings using pattern matching
2. **Classification**: Distinguish generic URLs from Git remote forms
3. **Validation**: Apply RFC 3986 syntax checks for generic URLs
4. **Transformation**: Convert Git remotes to canonical HTTPS form
5. **Normalization**: Apply host case, trimming, ...
6. **Deduplication**: Remove duplicates while preserving order
7. **Output**: Emit one URL per line

### Normalization Order
1. Extract candidate with boundary detection
2. Trim leading/trailing wrapper punctuation
3. Validate percent-encoding completeness
4. Classify as generic URL or Git remote
5. Apply transformation rules if Git remote
6. Normalize host (lowercase, punycode, port removal)
7. Deduplicate

## Test Requirements

### Unit Tests (UT)
- **UT-1**: RFC 3986 scheme validation
- **UT-2**: Host normalization (case, punycode, ports)
- **UT-3**: Wrapper punctuation trimming rules
- **UT-4**: Percent-encoding validation
- **UT-5**: Credential removal from userinfo

### Git Remote Tests (GRT)
- **GRT-1**: SCP-like parsing and transformation
- **GRT-2**: SSH scheme handling with ports and users
- **GRT-3**: Git protocol conversion
- **GRT-4**: HTTPS remote normalization
- **GRT-5**: Nested group/organization paths
- **GRT-6**: Edge cases (no .git, trailing slashes)

### Integration Tests (IT)
- **IT-1**: Mixed URL and remote input
- **IT-2**: Deduplication across different forms
- **IT-3**: Order preservation
- **IT-4**: Large input performance
- **IT-5**: Adversarial input handling

### Property-Based Tests (PBT)
- **PBT-1**: Normalization idempotency
- **PBT-2**: No credential leakage
- **PBT-3**: Valid output URL re-parsing
- **PBT-4**: Deterministic output

### Fuzz Testing (FUZZ)
- **FUZZ-1**: Arbitrary byte input safety
- **FUZZ-2**: No panic guarantee
- **FUZZ-3**: Performance regression detection

## Test Cases Catalog

### Basic URL Extraction
```
Input:  Visit https://example.com/path for more info.
Output: https://example.com/path
```

### Punctuation Trimming
```
Input:  Check out (https://github.com/user/repo).
Output: https://github.com/user/repo

Input:  Link: https://example.com/path.
Output: https://example.com/path

Input:  URL https://example.com/path...
Output: https://example.com/path...
```

### Git Remote Transformations
```
Input:  git@github.com:facebook/react.git
Output: https://github.com/facebook/react

Input:  ssh://git@gitlab.example.com:2222/group/project.git
Output: https://gitlab.example.com/group/project

Input:  git://kernel.org/pub/scm/git/git.git
Output: https://kernel.org/pub/scm/git/git

Input:  https://bitbucket.org/atlassian/stash.git
Output: https://bitbucket.org/atlassian/stash
```

### Credential Handling
```
Input:  https://user:token@api.github.com/repos
Output: https://api.github.com/repos

Input:  deploy@server.com:apps/myapp.git  
Output: https://server.com/apps/myapp
```

### Internationalization
```
Input:  https://测试.中国/path
Output: https://xn--0zwm56d.xn--fiqs8s/path
```

### Invalid Cases (No Output)
```
Input:  git@host:repo          # Missing slash
Input:  https://host/path%     # Incomplete percent-encoding
Input:  http://              # Missing host
```

### Deduplication
```
Input:  git@github.com:user/repo.git and https://github.com/user/repo
Output: https://github.com/user/repo
```

## Performance Benchmarks

### Benchmark Targets
- **B1**: 1MB mixed text, 0.5% URLs → < 30ms runtime
- **B2**: 10MB terminal buffer, 1% URLs → < 250ms runtime  
- **B3**: High density, 5% short URLs → ≤ 2 allocations per URL
- **B4**: Adversarial input (repeated colons) → linear time guarantee

### Memory Constraints
- Peak memory usage: < 4× input size
- No memory leaks on repeated invocations
- Garbage collection friendly (no long-lived references)

## Implementation Phases

### Phase 1: Foundation
- Core data structures and types
- Input reading and output writing
- Basic pattern matching framework

### Phase 2: URL Extraction  
- RFC 3986 validation
- Scheme and authority parsing
- Percent-encoding validation

### Phase 3: Git Remote Handling
- SCP-like pattern detection
- SSH/Git protocol parsing  
- Transformation to HTTPS canonical form

### Phase 4: Normalization
- Host case conversion and punycode
- Wrapper punctuation trimming
- Credential removal

### Phase 5: Integration
- Deduplication logic
- Order preservation
- Complete pipeline integration

### Phase 6: Testing & Validation
- Comprehensive test suite
- Fuzz testing harness
- Performance benchmarks

## Error Handling

### I/O Errors
- **Stdin read failure**: Exit with non-zero code
- **Stdout write failure**: Abort immediately
- **Partial operations**: Best effort completion

### Processing Errors  
- **Invalid UTF-8**: Skip or replace malformed sequences
- **Malformed URLs**: Exclude from output (no partial results)
- **Internal errors**: Exit non-zero with minimal error message

### Security Considerations
- Never output sensitive information
- No network side effects
- Bounded memory usage
- Safe regex patterns (RE2 compatible)

## Acceptance Criteria

### Functional Acceptance
- [ ] All RFC 3986 absolute URLs extracted correctly
- [ ] Git remote forms transformed to navigable HTTPS URLs
- [ ] Wrapper punctuation properly trimmed
- [ ] Credentials removed from all outputs
- [ ] Deduplication preserves first-occurrence order
- [ ] Deterministic output across runs

### Performance Acceptance  
- [ ] Processes 40+ MB/s on target hardware
- [ ] Linear time complexity on adversarial input
- [ ] Memory usage within specified bounds
- [ ] No performance regressions vs benchmarks

### Quality Acceptance
- [ ] Zero panics on fuzz testing
- [ ] 95%+ test coverage of normalization logic
- [ ] All benchmark targets met
- [ ] Property-based tests validate invariants

---

## Implementation Guidance

This specification provides the complete requirements for implementing urlpicker from scratch. The implementation should:

1. Follow the functional requirements exactly
2. Meet all performance benchmarks  
3. Pass the comprehensive test suite
4. Maintain security and privacy guarantees
5. Provide deterministic, reproducible behavior

Any ambiguities in this specification should be resolved by:
1. Consulting the referenced RFCs and official documentation
2. Choosing the most conservative/secure interpretation
3. Documenting the decision in implementation comments

The goal is a single-purpose, high-performance tool that reliably extracts and normalizes URLs from any text input while providing convenient Git remote transformations.
