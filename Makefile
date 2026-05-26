# 本地集成测试
GO ?= go

# 若存在 .env 则加载并 export 到 recipe（见 .env.example）
-include .env
export

# Makefile 变量与 .env 中 OPENAI_* 对齐（命令行 MODEL= / BASE_URL= 仍可覆盖）
MODEL ?= $(OPENAI_MODEL)
BASE_URL ?= $(OPENAI_BASE_URL)

CLI := $(GO) run ./cmd/mitmproxy2swagger

# testdata/xiezuo.har 在 .gitignore 中，需自行放置
XIEZUO_HAR := testdata/xiezuo.har
# 匹配 woa.wps.cn 下 /api/v1、v2、v3、v4；若只要 v2 可改为 https://woa.wps.cn/api/v2
XIEZUO_PREFIX := https://woa.wps.cn/api
XIEZUO_DIR := build/xiezuo
XIEZUO_SCHEMA := $(XIEZUO_DIR)/schema.yaml
XIEZUO_ENRICHED := $(XIEZUO_DIR)/enriched.yaml
XIEZUO_PROMPTS := $(XIEZUO_DIR)/prompts
XIEZUO_REDOC := $(XIEZUO_DIR)/redoc.html
SAMPLES := 5

REDOCLY ?= redocly

.PHONY: help xiezuo-pass1 xiezuo-pass2 xiezuo-curate-strip xiezuo-pass \
	xiezuo-enrich-prompts xiezuo-enrich xiezuo-redoc xiezuo-clean

help:
	@echo "xiezuo 工作流（依赖 $(XIEZUO_HAR)）："
	@echo "  1. make xiezuo-pass1          第一遍 Pass → $(XIEZUO_SCHEMA)"
	@echo "  2. 编辑 schema，去掉要生成的 path 上的 ignore:"
	@echo "  3. make xiezuo-pass2          第二遍 Pass，写入 paths"
	@echo "  4. make xiezuo-enrich-prompts 导出 prompt（不调 LLM）"
	@echo "     make xiezuo-enrich          真实 LLM enrich（需 OPENAI_API_KEY）"
	@echo "  5. make xiezuo-redoc           Redocly → $(XIEZUO_REDOC)"
	@echo ""
	@echo "可选:"
	@echo "  make xiezuo-curate-strip      批量去掉全部 ignore:（显式 opt-in，非默认）"
	@echo "  make xiezuo-pass              同 xiezuo-pass1"
	@echo "  make xiezuo-clean             删除 $(XIEZUO_DIR)"
	@echo "  变量: MODEL=...  BASE_URL=...  SAMPLES=1（或写入 .env，存在时自动加载）"
	@echo "        REDOCLY=redocly         （未安装时用: npx --yes @redocly/cli）"

# 兼容旧 target 名
xiezuo-pass: xiezuo-pass1

xiezuo-pass1: $(XIEZUO_SCHEMA)

$(XIEZUO_SCHEMA): $(XIEZUO_HAR)
	@test -f $(XIEZUO_HAR) || (echo "missing $(XIEZUO_HAR)" && exit 1)
	@mkdir -p $(XIEZUO_DIR)
	@echo "Pass 1: discover path templates (x-path-templates with ignore:)"
	$(CLI) pass -i $(XIEZUO_HAR) -o $(XIEZUO_SCHEMA) -p $(XIEZUO_PREFIX) -f har

xiezuo-curate-strip:
	@test -f $(XIEZUO_SCHEMA) || (echo "run make xiezuo-pass1 first" && exit 1)
	@echo "Strip all ignore: prefixes (opt-in; prefer manual curation)"
	$(GO) run ./cmd/stripignore -schema $(XIEZUO_SCHEMA)

xiezuo-pass2:
	@test -f $(XIEZUO_SCHEMA) || (echo "run make xiezuo-pass1 first" && exit 1)
	@echo "Pass 2: materialize curated paths"
	$(CLI) pass -i $(XIEZUO_HAR) -o $(XIEZUO_SCHEMA) -p $(XIEZUO_PREFIX) -f har

xiezuo-enrich-prompts:
	@test -f $(XIEZUO_SCHEMA) || (echo "run make xiezuo-pass1 first" && exit 1)
	@mkdir -p $(XIEZUO_PROMPTS)
	$(CLI) enrich \
		-i $(XIEZUO_HAR) \
		-s $(XIEZUO_SCHEMA) \
		-o $(XIEZUO_ENRICHED) \
		-p $(XIEZUO_PREFIX) \
		-f har \
		--samples $(or $(SAMPLES),1) \
		--emit-prompts $(XIEZUO_PROMPTS)

xiezuo-enrich:
	@test -f $(XIEZUO_SCHEMA) || (echo "run make xiezuo-pass1 first" && exit 1)
	@test -n "$$OPENAI_API_KEY" || (echo "set OPENAI_API_KEY" && exit 1)
	$(CLI) enrich \
		-i $(XIEZUO_HAR) \
		-s $(XIEZUO_SCHEMA) \
		-o $(XIEZUO_ENRICHED) \
		-p $(XIEZUO_PREFIX) \
		-f har \
		--samples $(or $(SAMPLES),1) \
		--force \
		$(if $(MODEL),--model $(MODEL),) \
		$(if $(BASE_URL),--base-url $(BASE_URL),)

xiezuo-redoc: $(XIEZUO_REDOC)

$(XIEZUO_REDOC): $(XIEZUO_ENRICHED)
	@test -f $(XIEZUO_ENRICHED) || (echo "missing $(XIEZUO_ENRICHED); run make xiezuo-enrich first" && exit 1)
	$(REDOCLY) build-docs $(XIEZUO_ENRICHED) -o $(XIEZUO_REDOC)

xiezuo-clean:
	rm -rf $(XIEZUO_DIR)
