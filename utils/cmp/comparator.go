package cmp

type Comparator interface {
	Compare(a, b []byte) int
}

//    <0 , if a < b
//    =0 , if a == b
//    >0 , if a > b
//type Comparator func(a, b interface{}) int
//
//func IntComparator(a, b interface{}) int {
//	aInt := a.(int)
//	bInt := b.(int)
//	return aInt - bInt
//}
