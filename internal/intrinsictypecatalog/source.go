// Package intrinsictypecatalog retains one fixed trusted publication. It does
// not register Types, allocate Resources, install defaults or admit operations.
package intrinsictypecatalog

import _ "embed"

//go:embed intrinsic-1.json
var selectedBytes []byte

// Source identifies checked source bytes; it is not a provider or ownership claim.
type Source struct {
	Repository string
	Commit     string
	Path       string
	SHA256     string
	Bytes      int
}

// SelectedSource identifies the qualified immutable source of this fixed loader.
func SelectedSource() Source {
	return Source{Repository: "gastownhall/bdp", Commit: "ba8d18c6355cbd6dd3499befe03641874bb01c73",
		Path: "specs/profiles/types/intrinsic-1.json", SHA256: "73b16c93b6963f9657be3569ba9d16f722edbf6f49090803e748c907c39aab7e", Bytes: 100125}
}

const namespace = "https://github.com/gastownhall/bdp/types/type-system/"
const schemaNamespace = "https://github.com/gastownhall/bdp/schemas/types/intrinsic-1/"
const dialect = "https://json-schema.org/draft/2020-12/schema"

// Pinned schema-value digests include the eight standard resources. The
// compiler's built-in reserved URI handling must not authenticate supplied data.
func selectedSchemaDigests() map[string]string {
	return map[string]string{
		"https://github.com/gastownhall/bdp/schemas/types/intrinsic-1/bead-type-definition.properties.schema.json": "139670760ca31dad57e8c37d5f2e19b4409bc8bdb88dd30bf525ac35361bd569",
		"https://github.com/gastownhall/bdp/schemas/types/intrinsic-1/link-type-definition.properties.schema.json": "86d8822466807f25d140d773dca0cc88ca7cedc3ff60b2f05bca428f23b6a1f6",
		"https://github.com/gastownhall/bdp/schemas/types/intrinsic-1/refines.properties.schema.json":              "4f02e74dfeeca30c7a3bdf6baa81a11385032b6174d4c05f7eb88fac6c31d2f7",
		"https://github.com/gastownhall/bdp/schemas/types/intrinsic-1/type-definition.properties.schema.json":      "de54fa5bac56907f71f3fb5101b0981420e89d35ac5d7e1ed91bf57c38b920cf",
		"https://json-schema.org/draft/2020-12/meta/applicator":                                                    "46a5141432ac51fff467e6c2230c19cba0b07d6a492c0fb062a1a852c370eed7",
		"https://json-schema.org/draft/2020-12/meta/content":                                                       "adcc808be6d10c12a2fcdf3123ecc4b56b18df58fc86001f2740ddd10fdeedfd",
		"https://json-schema.org/draft/2020-12/meta/core":                                                          "fc882f9fe8343beacb4b47b4a65441a9bfe2a15f279c66ceb6ddd65da7586edd",
		"https://json-schema.org/draft/2020-12/meta/format-annotation":                                             "f35165bc538974536af0c491819172f68e32e304e677742e67fe12f008731e60",
		"https://json-schema.org/draft/2020-12/meta/meta-data":                                                     "5b3b96d59a4a2f16fc72a0dcaf758aea40aae866a61ac9d0d71dd9d6a8603115",
		"https://json-schema.org/draft/2020-12/meta/unevaluated":                                                   "61f01418e98cdbb0ad0fbfbd593780bdadb7fb745a7adff646d26c103bc6d166",
		"https://json-schema.org/draft/2020-12/meta/validation":                                                    "34fba27426a2fe175b459326a4013434775f06dd33d1209cab9028decae1e240",
		"https://json-schema.org/draft/2020-12/schema":                                                             "c1cacf82bea665da4dbaf58eb341b7af39420d4c9248a012aba3436ef1f802bb",
	}
}
