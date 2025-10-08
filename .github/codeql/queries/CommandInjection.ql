/**
 * @name Command line injection via flag parsing
 * @description Detects potential command injection vulnerabilities in flag parsing functions
 * @kind path-problem
 * @problem.severity error
 * @security-severity 9.0
 * @precision high
 * @id go/flash-flags/command-injection
 * @tags security
 *       external/cwe/cwe-078
 *       external/cwe/cwe-088
 */

import go
import semmle.go.security.CommandInjection

/**
 * A source of untrusted data from command line flags
 */
class FlagSource extends Source {
  FlagSource() {
    exists(CallExpr call |
      call.getTarget().hasQualifiedName("", "Parse") or
      call.getTarget().hasQualifiedName("", "ParseStringSlice") or
      call.getTarget().hasQualifiedName("", "LoadConfig")
    |
      this = call.getAnArgument()
    )
  }
}

/**
 * A sink where command injection could occur
 */
class CommandSink extends Sink {
  CommandSink() {
    exists(CallExpr call |
      call.getTarget().hasQualifiedName("os/exec", "Command") or
      call.getTarget().hasQualifiedName("os/exec", "CommandContext") or
      call.getTarget().hasQualifiedName("syscall", "Exec") or
      call.getTarget().hasQualifiedName("syscall", "ForkExec")
    |
      this = call.getAnArgument()
    )
  }
}

/**
 * Security validation that prevents command injection
 */
class SecurityValidation extends Sanitizer {
  SecurityValidation() {
    exists(CallExpr call |
      call.getTarget().hasQualifiedName("", "validateSecurityConstraints")
    |
      this = call
    )
  }
}

from FlagSource source, CommandSink sink
where TaintTracking::flow(source, sink) and
      not exists(SecurityValidation validation |
        TaintTracking::flow(source, validation) and
        TaintTracking::flow(validation, sink))
select sink, source, sink, "Command injection vulnerability: untrusted flag data $@ reaches command execution $@ without security validation",
       source, "source", sink, "sink"