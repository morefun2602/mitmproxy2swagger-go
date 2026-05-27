# 本地集成测试
GO ?= go

# 若存在 .env 则加载并 export 到 recipe（见 .env.example）
-include .env
export

# Makefile 变量与 .env 中 OPENAI_* 对齐（命令行 MODEL= / BASE_URL= 仍可覆盖）
MODEL ?= $(OPENAI_MODEL)
BASE_URL ?= $(OPENAI_BASE_URL)

CLI := $(GO) run ./cmd/mitmproxy2swagger

# 示例工作流：使用仓库内公开 HAR；私有 capture 请放到 testdata/local/（已 gitignore）
EXAMPLE_HAR := testdata/captures/im.har
EXAMPLE_PREFIX := https://woa.wps.cn/api
EXAMPLE_DIR := build/example
EXAMPLE_SCHEMA := $(EXAMPLE_DIR)/schema.yaml
EXAMPLE_ENRICHED := $(EXAMPLE_DIR)/enriched.yaml
EXAMPLE_PROMPTS := $(EXAMPLE_DIR)/prompts
EXAMPLE_REDOC := $(EXAMPLE_DIR)/redoc.html
EXAMPLE_SUGGESTIONS := $(EXAMPLE_DIR)/template-suggestions.yaml
EXAMPLE_AUTH_OBS := $(EXAMPLE_DIR)/auth-observations.yaml
EXAMPLE_TAGS := $(EXAMPLE_DIR)/tags.yaml
# 增量更新：仅录缺失接口的 HAR（见 README.zh-CN.md「增量更新」）
INCREMENTAL_HAR ?= testdata/local/incremental.har
SAMPLES := 5
CONCURRENCY ?= 10

REDOCLY ?= redocly

.PHONY: help example-pass1 example-curate-auto example-auth-observe example-auth-apply \
	example-curate-suggest-prompts example-curate-suggest example-curate-apply \
	example-pass2 example-enrich-prompts example-enrich example-tags-apply example-redoc example-clean \
	example-incremental-pass1 example-incremental-pass2 example-incremental-enrich

help:
	@echo "示例工作流（依赖 $(EXAMPLE_HAR)）："
	@echo "  1.  make example-pass1                 第一遍 Pass → $(EXAMPLE_SCHEMA)"
	@echo "  2.  make example-curate-auto           聚类 x-path-templates"
	@echo "  3.  make example-auth-observe          鉴权观测 → $(EXAMPLE_AUTH_OBS)"
	@echo "  4.  编辑 $(EXAMPLE_AUTH_OBS)，填 verified / combination"
	@echo "  5.  make example-auth-apply            写入 securitySchemes"
	@echo "  6.  make example-curate-suggest-prompts  导出 curate LLM prompt（不调 API，可选）"
	@echo "  7.  make example-curate-suggest        LLM 合并建议 → $(EXAMPLE_SUGGESTIONS)（可选，需 OPENAI_API_KEY）"
	@echo "  8.  make example-curate-apply          应用 template-suggestions（可选）"
	@echo "  9.  编辑 $(EXAMPLE_SCHEMA)，去掉要生成的 path 上的 ignore:"
	@echo "  10. make example-pass2                 第二遍 Pass，写入 paths"
	@echo "  11. make example-enrich-prompts        导出 enrich prompt（不调 LLM）"
	@echo "  12. make example-enrich                LLM enrich → $(EXAMPLE_ENRICHED)（需 OPENAI_API_KEY）"
	@echo "  13. make example-tags-apply            应用 $(EXAMPLE_TAGS) 分组 → $(EXAMPLE_ENRICHED)"
	@echo "  14. make example-redoc                 Redocly → $(EXAMPLE_REDOC)"
	@echo ""
	@echo "增量更新（需已有 $(EXAMPLE_ENRICHED)，HAR=$(INCREMENTAL_HAR)）："
	@echo "  make example-incremental-pass1         Pass 1 → $(EXAMPLE_ENRICHED)"
	@echo "  编辑 $(EXAMPLE_ENRICHED)：仅对新增 x-path-templates 去掉 ignore:"
	@echo "  make example-incremental-pass2         Pass 2"
	@echo "  make example-incremental-enrich      enrich（无 --force）"
	@echo ""
	@echo "其它:"
	@echo "  make example-clean                     删除 $(EXAMPLE_DIR)"
	@echo "  变量: MODEL=...  BASE_URL=...  SAMPLES=1  CONCURRENCY=10（或写入 .env，存在时自动加载）"
	@echo "        INCREMENTAL_HAR=...              增量 HAR 路径"
	@echo "        REDOCLY=redocly                  （未安装时用: npx --yes @redocly/cli）"


example-pass1: $(EXAMPLE_SCHEMA)

$(EXAMPLE_SCHEMA): $(EXAMPLE_HAR)
	@test -f $(EXAMPLE_HAR) || (echo "missing $(EXAMPLE_HAR)" && exit 1)
	@mkdir -p $(EXAMPLE_DIR)
	@echo "Pass 1: discover path templates (x-path-templates with ignore:)"
	$(CLI) pass -i $(EXAMPLE_HAR) -o $(EXAMPLE_SCHEMA) -p $(EXAMPLE_PREFIX) -f har

example-curate-auto:
	@test -f $(EXAMPLE_SCHEMA) || (echo "run make example-pass1 first" && exit 1)
	@echo "Curate --auto: cluster x-path-templates"
	$(CLI) curate --auto -o $(EXAMPLE_SCHEMA)

example-auth-observe:
	@test -f $(EXAMPLE_HAR) || (echo "missing $(EXAMPLE_HAR)" && exit 1)
	@mkdir -p $(EXAMPLE_DIR)
	@echo "Auth observe: credential statistics"
	$(CLI) auth observe -i $(EXAMPLE_HAR) -p $(EXAMPLE_PREFIX) -f har -o $(EXAMPLE_AUTH_OBS)

example-auth-apply:
	@test -f $(EXAMPLE_SCHEMA) || (echo "run make example-pass1 first" && exit 1)
	@test -f $(EXAMPLE_AUTH_OBS) || (echo "run make example-auth-observe first" && exit 1)
	@echo "Auth apply: securitySchemes + security from $(EXAMPLE_AUTH_OBS)"
	$(CLI) auth apply -s $(EXAMPLE_SCHEMA) --observations $(EXAMPLE_AUTH_OBS)

example-curate-suggest-prompts:
	@test -f $(EXAMPLE_SCHEMA) || (echo "run make example-pass1 first" && exit 1)
	@test -f $(EXAMPLE_HAR) || (echo "missing $(EXAMPLE_HAR)" && exit 1)
	@mkdir -p $(EXAMPLE_DIR)/curate-prompts
	$(CLI) curate --llm-suggest -o $(EXAMPLE_SCHEMA) -i $(EXAMPLE_HAR) -p $(EXAMPLE_PREFIX) -f har \
		--emit-prompts $(EXAMPLE_DIR)/curate-prompts

example-curate-suggest:
	@test -f $(EXAMPLE_SCHEMA) || (echo "run make example-pass1 first" && exit 1)
	@test -f $(EXAMPLE_HAR) || (echo "missing $(EXAMPLE_HAR)" && exit 1)
	@test -n "$$OPENAI_API_KEY" || (echo "set OPENAI_API_KEY" && exit 1)
	@echo "Curate --llm-suggest: template merge suggestions"
	$(CLI) curate --llm-suggest -o $(EXAMPLE_SCHEMA) -i $(EXAMPLE_HAR) -p $(EXAMPLE_PREFIX) -f har \
		--suggestions-out $(EXAMPLE_SUGGESTIONS) \
		$(if $(MODEL),--model $(MODEL),) \
		$(if $(BASE_URL),--base-url $(BASE_URL),)

example-curate-apply:
	@test -f $(EXAMPLE_SCHEMA) || (echo "run make example-pass1 first" && exit 1)
	@test -f $(EXAMPLE_SUGGESTIONS) || (echo "missing $(EXAMPLE_SUGGESTIONS); run make example-curate-suggest" && exit 1)
	@echo "Curate --apply-suggestions"
	$(CLI) curate -o $(EXAMPLE_SCHEMA) --apply-suggestions $(EXAMPLE_SUGGESTIONS)

example-pass2:
	@test -f $(EXAMPLE_SCHEMA) || (echo "run make example-pass1 first" && exit 1)
	@echo "Pass 2: materialize curated paths"
	$(CLI) pass -i $(EXAMPLE_HAR) -o $(EXAMPLE_SCHEMA) -p $(EXAMPLE_PREFIX) -f har

example-enrich-prompts:
	@test -f $(EXAMPLE_SCHEMA) || (echo "run make example-pass1 first" && exit 1)
	@mkdir -p $(EXAMPLE_PROMPTS)
	$(CLI) enrich \
		-i $(EXAMPLE_HAR) \
		-s $(EXAMPLE_SCHEMA) \
		-o $(EXAMPLE_ENRICHED) \
		-p $(EXAMPLE_PREFIX) \
		-f har \
		--samples $(or $(SAMPLES),1) \
		--emit-prompts $(EXAMPLE_PROMPTS)

example-enrich:
	@test -f $(EXAMPLE_SCHEMA) || (echo "run make example-pass1 first" && exit 1)
	@test -n "$$OPENAI_API_KEY" || (echo "set OPENAI_API_KEY" && exit 1)
	$(CLI) enrich \
		-i $(EXAMPLE_HAR) \
		-s $(EXAMPLE_SCHEMA) \
		-o $(EXAMPLE_ENRICHED) \
		-p $(EXAMPLE_PREFIX) \
		-f har \
		--samples $(or $(SAMPLES),1) \
		--force \
		--concurrency $(or $(CONCURRENCY),10) \
		$(if $(MODEL),--model $(MODEL),) \
		$(if $(BASE_URL),--base-url $(BASE_URL),)

example-tags-apply:
	@test -f $(EXAMPLE_ENRICHED) || (echo "missing $(EXAMPLE_ENRICHED); run make example-enrich first" && exit 1)
	@test -f $(EXAMPLE_TAGS) || (echo "missing $(EXAMPLE_TAGS)" && exit 1)
	@echo "Tags apply: $(EXAMPLE_TAGS) → $(EXAMPLE_ENRICHED)"
	$(CLI) tags apply -s $(EXAMPLE_ENRICHED) -t $(EXAMPLE_TAGS)

example-redoc: $(EXAMPLE_REDOC)

$(EXAMPLE_REDOC): $(EXAMPLE_ENRICHED) $(EXAMPLE_TAGS)
	@test -f $(EXAMPLE_ENRICHED) || (echo "missing $(EXAMPLE_ENRICHED); run make example-enrich first" && exit 1)
	@$(MAKE) example-tags-apply
	$(REDOCLY) build-docs $(EXAMPLE_ENRICHED) -o $(EXAMPLE_REDOC)

example-incremental-pass1:
	@test -f $(EXAMPLE_ENRICHED) || (echo "missing $(EXAMPLE_ENRICHED); complete initial enrich workflow first" && exit 1)
	@test -f $(INCREMENTAL_HAR) || (echo "missing $(INCREMENTAL_HAR); export incremental HAR (see README.zh-CN.md)" && exit 1)
	@echo "Incremental pass 1: discover new templates into $(EXAMPLE_ENRICHED)"
	$(CLI) pass -i $(INCREMENTAL_HAR) -o $(EXAMPLE_ENRICHED) -p $(EXAMPLE_PREFIX) -f har

example-incremental-pass2:
	@test -f $(EXAMPLE_ENRICHED) || (echo "missing $(EXAMPLE_ENRICHED)" && exit 1)
	@test -f $(INCREMENTAL_HAR) || (echo "missing $(INCREMENTAL_HAR)" && exit 1)
	@echo "Incremental pass 2: materialize new paths (curate new ignore: entries first)"
	$(CLI) pass -i $(INCREMENTAL_HAR) -o $(EXAMPLE_ENRICHED) -p $(EXAMPLE_PREFIX) -f har

example-incremental-enrich:
	@test -f $(EXAMPLE_ENRICHED) || (echo "missing $(EXAMPLE_ENRICHED)" && exit 1)
	@test -f $(INCREMENTAL_HAR) || (echo "missing $(INCREMENTAL_HAR)" && exit 1)
	@test -n "$$OPENAI_API_KEY" || (echo "set OPENAI_API_KEY" && exit 1)
	@echo "Incremental enrich: fill empty semantic fields only (no --force)"
	$(CLI) enrich \
		-i $(INCREMENTAL_HAR) \
		-s $(EXAMPLE_ENRICHED) \
		-o $(EXAMPLE_ENRICHED) \
		-p $(EXAMPLE_PREFIX) \
		-f har \
		--samples $(or $(SAMPLES),1) \
		--concurrency $(or $(CONCURRENCY),10) \
		$(if $(MODEL),--model $(MODEL),) \
		$(if $(BASE_URL),--base-url $(BASE_URL),)

example-clean:
	rm -rf $(EXAMPLE_DIR)
