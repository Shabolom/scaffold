package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	moduleName string
	force      bool
	withPG     bool
	withKafka  bool
)

var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Создать структуру проекта",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		mod := moduleName
		if mod == "" {
			mod = projectName
		}

		return createProject(projectName, mod, force, withPG, withKafka)
	},
}

func init() {
	createCmd.Flags().StringVar(
		&moduleName,
		"module",
		"",
		"имя go-модуля, например github.com/user/project",
	)

	createCmd.Flags().BoolVar(
		&force,
		"force",
		false,
		"разрешить создание, даже если папка уже существует",
	)

	createCmd.Flags().BoolVar(
		&withPG,
		"postgres",
		false,
		"добавить конфиг и docker-compose для PostgreSQL",
	)

	createCmd.Flags().BoolVar(
		&withKafka,
		"kafka",
		false,
		"добавить конфиг и docker-compose для Kafka",
	)

	rootCmd.AddCommand(createCmd)
}

func createProject(projectName, moduleName string, force, withPG, withKafka bool) error {
	root := projectName

	if !force {
		if _, err := os.Stat(root); err == nil {
			return fmt.Errorf("папка проекта %q уже существует", root)
		}
	}

	dirs := []string{
		filepath.Join(root, "build", "app", "migrations"),
		filepath.Join(root, "build", "local"),
		filepath.Join(root, "cmd", projectName),
		filepath.Join(root, "docks"),

		filepath.Join(root, "internal", "config"),
		filepath.Join(root, "internal", "constrain"),
		filepath.Join(root, "internal", "di"),
		filepath.Join(root, "internal", "dto"),
		filepath.Join(root, "internal", "render"),
		filepath.Join(root, "internal", "handler"),
		filepath.Join(root, "internal", "service"),
		filepath.Join(root, "internal", "repository"),

		filepath.Join(root, "pkg", "shortcut"),
		filepath.Join(root, "pkg", "utils"),
	}

	files := map[string]string{
		filepath.Join(root, "go.mod"): fmt.Sprintf(`module %s

go 1.25.0
`, moduleName),

		filepath.Join(root, "cmd", projectName, "main.go"): fmt.Sprintf(`package main

import "fmt"

func main() {
	fmt.Println("hello from %s")
}
`, projectName),

		filepath.Join(root, "README.md"): fmt.Sprintf("# %s\n", projectName),
		filepath.Join(root, ".env"):      buildEnv(projectName, withPG, withKafka),
	}

	if withPG || withKafka {
		files[filepath.Join(root, "docker-compose.yml")] = buildDockerCompose(projectName, withPG, withKafka)
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("ошибка создания папки %s: %w", dir, err)
		}
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("ошибка создания файла %s: %w", path, err)
		}
	}

	fmt.Printf("проект создан: %s\n", root)
	return nil
}

func buildEnv(projectName string, withPG, withKafka bool) string {
	var b strings.Builder

	b.WriteString("APP_NAME=" + projectName + "\n")
	b.WriteString("APP_PORT=8080\n")

	if withPG {
		b.WriteString("\n")
		b.WriteString("POSTGRES_HOST=localhost\n")
		b.WriteString("POSTGRES_PORT=5432\n")
		b.WriteString("POSTGRES_DB=postgres\n")
		b.WriteString("POSTGRES_USER=postgres\n")
		b.WriteString("POSTGRES_PASSWORD=postgres\n")
		b.WriteString("POSTGRES_SSLMODE=disable\n")
		b.WriteString("POSTGRES_DSN=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable\n")
	}

	if withKafka {
		b.WriteString("\n")
		b.WriteString("KAFKA_BROKERS=localhost:9092\n")
		b.WriteString("KAFKA_TOPIC=example-topic\n")
		b.WriteString("KAFKA_GROUP_ID=" + projectName + "-group\n")
	}

	return b.String()
}

func buildDockerCompose(projectName string, withPG, withKafka bool) string {
	var b strings.Builder

	b.WriteString("version: '3.9'\n\n")
	b.WriteString("services:\n")

	if withPG {
		b.WriteString(`  postgres:
    image: postgres:16
    container_name: ` + projectName + `-postgres
    restart: always
    environment:
      POSTGRES_DB: postgres
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

`)
	}

	if withKafka {
		b.WriteString(`  kafka:
    image: apache/kafka:4.1.2
    container_name: ` + projectName + `-kafka
    restart: always
    ports:
      - "9092:9092"
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_PROCESS_ROLES: broker,controller
      KAFKA_LISTENERS: PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT
      KAFKA_CONTROLLER_QUORUM_VOTERS: 1@localhost:9093
      KAFKA_INTER_BROKER_LISTENER_NAME: PLAINTEXT
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1
      KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS: 0
      CLUSTER_ID: MkU3OEVBNTcwNTJENDM2Qk

  kafka-ui:
    image: provectuslabs/kafka-ui:latest
    container_name: ` + projectName + `-kafka-ui
    restart: always
    ports:
      - "8081:8080"
    environment:
      KAFKA_CLUSTERS_0_NAME: local
      KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS: kafka:9092
    depends_on:
      - kafka

`)
	}

	if withPG {
		b.WriteString("volumes:\n")
		b.WriteString("  postgres_data:\n")
	}

	return b.String()
}
