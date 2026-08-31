package algo

import "testing"

func TestTwoSum(t *testing.T) {
	nums := []int{2, 4, 5, 6, 7, 8, 10, 15, 3, 7}
	target := 20
	k1, k2, ok := twoSum(nums, target)
	if !ok {
		t.Error(k1, k2)
	}
	t.Log(k1, k2)
}
