## Skills

Skills — это локальные инструкции и справочные материалы, которые агент может использовать как переиспользуемые модули поведения.

Обычно skill — это папка, в которой:
- `SKILL.md` — точка входа с описанием skill
- дополнительные `.md` файлы — вспомогательные инструкции, примеры и правила
- вложенные каталоги — разбиение по слоям, темам или сценариям

Пример:

```text
skills/
└── crud/
    ├── SKILL.md
    └── layers/
        ├── domain/guide.md
        ├── handler/guide.md
        ├── model/guide.md
        ├── repo/guide.md
        └── usecase/guide.md
```

В этой схеме:
- `SKILL.md` описывает назначение skill и служит основной точкой входа
- остальные файлы помогают агенту применять skill более точно и последовательно
- каждый skill хранится в отдельной директории внутри `skills/`

## Установка

Локальные skills из `.aiassistant/skills/` можно скопировать в директории, откуда их читает соответствующий инструмент.

или

```bash
npx skills add ./.aiasstant/skills/
```

### Codex

```bash
mkdir -p ~/.agents/skills
cp -r .aiassistant/skills/* ~/.agents/skills/
```

### Claude Code

```bash
mkdir -p ~/.claude/skills
cp -r .aiassistant/skills/* ~/.claude/skills/
```
