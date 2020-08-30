package writer

import "sort"
import "fmt"

// Empty!

func sortedKeys(m map[string][]string) (ret []string) {
	ret = make([]string,0,len(m))
	for k := range m { ret = append(ret,k) }
	sort.Strings(ret)
	return
}

func debug(i ...interface{}) {
	fmt.Println(i...)
}

