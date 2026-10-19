package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadUsesDevelopmentDefaults(t *testing.T) {
	cfg, err := load(func(string) string { return "" }, filepath.Join(t.TempDir(), "missing.env"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.Difficulty != 16 || cfg.InitialSupply != 1000 {
		t.Fatalf("defaults = %+v", cfg)
	}
	if cfg.BlockInterval != 2*time.Second {
		t.Fatalf("intervalo = %s", cfg.BlockInterval)
	}
	if cfg.ListenAddr() != ":8080" {
		t.Fatalf("listen = %s", cfg.ListenAddr())
	}
	if len(cfg.AllowedOrigins) == 0 {
		t.Fatal("se esperaban orígenes de desarrollo")
	}
}

func TestLoadPrefersEnvironmentOverFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "PORT=9090\nDIFFICULTY=8\nINITIAL_SUPPLY=5000\nALLOWED_ORIGINS=\"http://app.test\"\nBLOCK_INTERVAL_MS=250\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := load(func(key string) string {
		if key == "PORT" {
			return "7070"
		}
		return ""
	}, path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "7070" {
		t.Fatalf("PORT = %s", cfg.Port)
	}
	if cfg.Difficulty != 8 || cfg.InitialSupply != 5000 {
		t.Fatalf("archivo = %+v", cfg)
	}
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "http://app.test" {
		t.Fatalf("orígenes = %v", cfg.AllowedOrigins)
	}
	if cfg.BlockInterval != 250*time.Millisecond {
		t.Fatalf("intervalo = %s", cfg.BlockInterval)
	}
}

func TestLoadRejectsInvalidDifficulty(t *testing.T) {
	_, err := load(func(key string) string {
		if key == "DIFFICULTY" {
			return "no-es-un-numero"
		}
		return ""
	}, filepath.Join(t.TempDir(), "missing.env"))
	if err == nil {
		t.Fatal("se esperaba un error de configuración")
	}
}
