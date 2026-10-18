package i18n

import (
	"reflect"
	"regexp"
	"testing"
)

// verbRE matches a printf verb such as %s, %d or %v.
var verbRE = regexp.MustCompile(`%[a-zA-Z]`)

// TestTablesConsistent guards against a translation being added to only one
// language or with a different number/order of printf verbs.
func TestTablesConsistent(t *testing.T) {
	zh, en := tables[ZH], tables[EN]
	for key := range zh {
		if _, ok := en[key]; !ok {
			t.Errorf("key %q is missing from EN", key)
		}
	}
	for key := range en {
		if _, ok := zh[key]; !ok {
			t.Errorf("key %q is missing from ZH", key)
		}
	}
	for key, z := range zh {
		e, ok := en[key]
		if !ok {
			continue
		}
		if zv, ev := verbRE.FindAllString(z, -1), verbRE.FindAllString(e, -1); !reflect.DeepEqual(zv, ev) {
			t.Errorf("key %q verb mismatch: ZH %v, EN %v", key, zv, ev)
		}
	}
}
