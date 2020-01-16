package util

func Truncate(str string, num int) string {
	result := str
	if len(str) > num {
		result = str[0:num]
	}
	return result
}
