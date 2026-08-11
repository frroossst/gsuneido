// External test package: builtin imports typeinfer, which imports engine, so
// an in-package test could not import builtin back.
package engine_test

// builtin's init seeds the checker's builtin signature database from
// gsuneido's own registrations - see loadTypeCheckerAnnotations in
// builtin/typechecker.go. These tests need it loaded.
import _ "github.com/apmckinlay/gsuneido/builtin"
