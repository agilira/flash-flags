/**
 * @name Buffer overflow in flag parsing
 * @description Detects potential buffer overflow vulnerabilities due to insufficient input validation
 * @kind problem
 * @problem.severity error
 * @security-severity 8.0
 * @precision medium
 * @id go/flash-flags/buffer-overflow
 * @tags security
 *       external/cwe/cwe-120
 *       external/cwe/cwe-121
 */

import go

/**
 * Flag parsing functions that handle user input
 */
class FlagParsingFunction extends Function {
  FlagParsingFunction() {
    this.hasQualifiedName("", "Parse") or
    this.hasQualifiedName("", "ParseStringSlice") or
    this.hasQualifiedName("", "setFlagValue") or
    this.hasQualifiedName("", "LoadConfig")
  }
}

/**
 * Check if a string parameter has length validation
 */
predicate hasLengthValidation(Parameter p) {
  exists(CallExpr lengthCheck, ComparisonExpr comp |
    lengthCheck.getTarget().hasQualifiedName("", "len") and
    lengthCheck.getAnArgument() = p.getAReference() and
    comp.getAnOperand() = lengthCheck and
    (comp.getOperator() = "<" or comp.getOperator() = ">" or 
     comp.getOperator() = "<=" or comp.getOperator() = ">=")
  )
}

/**
 * Check if input goes through security validation
 */
predicate hasSecurityValidation(Parameter p) {
  exists(CallExpr validation |
    validation.getTarget().hasQualifiedName("", "validateSecurityConstraints") and
    validation.getAnArgument() = p.getAReference()
  )
}

from FlagParsingFunction f, Parameter p
where p = f.getAParameter() and
      p.getType().toString().matches("%string%") and
      not hasLengthValidation(p) and
      not hasSecurityValidation(p)
select f, "Function $@ may be vulnerable to buffer overflow: parameter $@ lacks proper input validation",
       f, f.getName(), p, p.getName()