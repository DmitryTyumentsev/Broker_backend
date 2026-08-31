package algo

func twoSum(nums []int, target int) (int, int, bool) {
	for k1, v1 := range nums {
		for k2, v2 := range nums {
			if target == v1+v2 {
				return k1, k2, true
			}
		}
	}
	return 0, 0, false
}
