/**
 * @name Path traversal via config loading
 * @description Detects path traversal vulnerabilities in config file loading
 * @kind path-problem  
 * @problem.severity error
 * @security-severity 8.5
 * @precision high
 * @id go/flash-flags/path-traversal
 * @tags security
 *       external/cwe/cwe-022
 *       external/cwe/cwe-023
 */

import go
import semmle.go.security.PathInjection

/**
 * Source of potentially malicious file paths from config loading
 */
class ConfigPathSource extends Source {
  ConfigPathSource() {
    exists(CallExpr call |
      call.getTarget().hasQualifiedName("", "LoadConfig") and
      this = call.getAnArgument()
    ) or
    exists(StringLiteral s |
      s.getValue().regexpMatch(".*\\.\\./.*") and
      this = s
    )
  }
}

/**
 * File operations that could be vulnerable to path traversal
 */
class FileOperationSink extends Sink {
  FileOperationSink() {
    exists(CallExpr call |
      call.getTarget().hasQualifiedName("os", "Open") or
      call.getTarget().hasQualifiedName("os", "OpenFile") or
      call.getTarget().hasQualifiedName("ioutil", "ReadFile") or
      call.getTarget().hasQualifiedName("io/ioutil", "ReadFile") or
      call.getTarget().hasQualifiedName("os", "ReadFile")
    |
      this = call.getAnArgument(0)
    )
  }
}

/**
 * Path sanitization that prevents traversal attacks
 */
class PathSanitizer extends Sanitizer {
  PathSanitizer() {
    exists(CallExpr call |
      call.getTarget().hasQualifiedName("filepath", "Clean") or
      call.getTarget().hasQualifiedName("filepath", "Abs") or
      call.getTarget().hasQualifiedName("", "validateSecurityConstraints")
    |
      this = call
    )
  }
}

from ConfigPathSource source, FileOperationSink sink
where TaintTracking::flow(source, sink) and
      not exists(PathSanitizer sanitizer |
        TaintTracking::flow(source, sanitizer) and
        TaintTracking::flow(sanitizer, sink))
select sink, source, sink, "Path traversal vulnerability: config path $@ reaches file operation $@ without proper sanitization",
       source, "source", sink, "sink"