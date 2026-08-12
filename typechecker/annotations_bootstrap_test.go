// external test package - builtin imports typechecker, so we can't import it back
package typechecker_test

// builtin's init loads the signature database - see loadTypeCheckerAnnotations
import _ "github.com/apmckinlay/gsuneido/builtin"
