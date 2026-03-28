package vdf

func mergeMaps(dst, src map[string]any) {
	type pair struct {
		dst map[string]any
		src map[string]any
	}
	stack := []pair{{dst, src}}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		d, s := current.dst, current.src
		for srcKey, srcVal := range s {
			if dstKey, exists := d[srcKey]; exists {
				dstMap, ok1 := dstKey.(map[string]any)
				srcMap, ok2 := srcVal.(map[string]any)

				if ok1 && ok2 {
					stack = append(stack, pair{dstMap, srcMap})
				} else {
					d[srcKey] = srcVal
				}
			} else {
				d[srcKey] = srcVal
			}
		}
	}
}
