package config

import (
	"os"
	"path/filepath"
	"testing"
)

// ChdirToRepoRoot переводит тест в корень репозитория.
//
// Файл называется testhelper.go, а не *_for_test.go: суффикс _test.go
// делает файл видимым только внутри тестов СВОЕГО пакета, а хелпер нужен
// тестам конфигов всех трёх сервисов.
//
// Зачем: viper ищет configs/<env>.yaml по путям, относительным от рабочего
// каталога процесса, а `go test` запускает тест из каталога пакета.
// Без этого тест на загрузку настоящих конфигов проверял бы только то,
// что файл не нашёлся.
//
// t.Chdir сам вернёт каталог обратно после теста.
func ChdirToRepoRoot(tb testing.TB) {
	tb.Helper()

	dir, err := os.Getwd()
	if err != nil {
		tb.Fatalf("getwd: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			tb.Chdir(dir)

			return
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			tb.Fatalf("go.mod не найден выше %s", dir)
		}

		dir = parent
	}
}
