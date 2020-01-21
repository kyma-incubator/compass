package util

import (
	"fmt"
	"strings"
)

// ClusterName creates from given string another string that matches Gardener cluster naming convention
// Requirements are strict:
// Cluster.Name must
//      - start with a lowercase letter
//  	- up to 19 lowercase letters, numbers, or hyphens,
//		- cannot end with a hyphen
//
// and since Gardener adds word 'shoot' to name the final length is 14
func ClusterName(str string) string {
	name := strings.ReplaceAll(str, "-", "")
	name = fmt.Sprintf("%.14s", name)
	name = strings.ToLower(name)
	name = strings.Replace(name, string(name[0]), "f", 1)
	return name
}
