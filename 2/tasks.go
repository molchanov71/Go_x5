package homework

// PairInput соответствует объекту nums/target из условия контеста.
type PairInput struct {
	Nums   []int
	Target int
}

// longestPrefix возвращает самый длинный общий префикс строк.
func longestPrefix(words []string) string {
	if words == nil || len(words) == 0 {
		return ""
	}
	res := ""
	for i := 0; ; i++ {
		if i >= len(words[0]) {
			break
		}
		r := []rune(words[0])[i]
		correct := true
		for _, word := range words[1:] {
			if i >= len(word) || []rune(word)[i] != r {
				correct = false
				break
			}
		}
		if correct {
			res = res + string(r)
		} else {
			break
		}
	}
	return res
}

// pairIndices возвращает индексы двух разных элементов по возрастанию.
func pairIndices(input PairInput) [2]int {
	for i := 0; i < len(input.Nums)-1; i++ {
		toSearch := input.Target - input.Nums[i]
		for j := i + 1; j < len(input.Nums); j++ {
			if toSearch == input.Nums[j] {
				return [2]int{i, j}
			}
		}
	}
	return [2]int{-1, -1}
}

// hasDuplicate сообщает, встречается ли какое-либо значение повторно.
func hasDuplicate(nums []int) bool {
	if nums == nil || len(nums) == 0 || len(nums) == 1 {
		return false
	}
	freq := make(map[int]struct{}, len(nums))
	for _, num := range nums {
		if _, exists := freq[num]; !exists {
			freq[num] = struct{}{}
		} else {
			return true
		}
	}
	return false
}

// compactSorted создаёт независимый слайс уникальных значений.
func compactSorted(nums []int) []int {
	if nums == nil || len(nums) == 0 {
		return []int{}
	}
	var res []int
	for _, num := range nums {
		if len(res) == 0 || res[len(res)-1] != num {
			res = append(res, num)
		}
	}
	return res
}

// balancedBrackets проверяет пары и порядок закрытия скобок.
func balancedBrackets(text string) bool {
	var stack []rune
	for _, r := range text {
		switch r {
		case '(', '[', '{':
			stack = append(stack, r)
		case ')':
			if len(stack) == 0 || stack[len(stack)-1] != '(' {
				return false
			}
			stack = stack[:len(stack)-1]
		case ']':
			if len(stack) == 0 || stack[len(stack)-1] != '[' {
				return false
			}
			stack = stack[:len(stack)-1]
		case '}':
			if len(stack) == 0 || stack[len(stack)-1] != '{' {
				return false
			}
			stack = stack[:len(stack)-1]
		default:
			return false
		}
	}
	if len(stack) > 0 {
		return false
	}
	return true
}
