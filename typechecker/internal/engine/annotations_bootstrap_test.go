// external test package - builtin imports engine (via typechecker), so we
// can't import it back
package engine_test

// builtin's init loads the signature database - see loadTypeCheckerAnnotations
import _ "github.com/apmckinlay/gsuneido/builtin"
