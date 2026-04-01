package sdk

import "fmt"

// buildQuery 는 ListOpts 를 쿼리 문자열로 변환한다
func buildQuery(opts *ListOpts) string {
	if opts == nil {
		return ""
	}

	q := ""
	sep := "?"

	if opts.Page > 0 {
		q += fmt.Sprintf("%spage=%d", sep, opts.Page)
		sep = "&"
	}
	if opts.Size > 0 {
		q += fmt.Sprintf("%ssize=%d", sep, opts.Size)
		sep = "&"
	}
	for k, v := range opts.Filter {
		q += fmt.Sprintf("%s%s=%s", sep, k, v)
		sep = "&"
	}

	return q
}
