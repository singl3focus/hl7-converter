# План улучшений

Этот документ фиксирует практический план развития репозитория `hl7-converter`. План разбит по горизонту выполнения, чтобы можно было отдельно вести быстрые улучшения, структурную доработку и долгосрочное развитие проекта.

## Цели

- Повысить доверие к репозиторию как к поддерживаемому Go-пакету.
- Сделать релизы предсказуемыми и видимыми через GitHub Releases.
- Упростить понимание проекта только по README, без чтения кода.
- Убрать из корня репозитория тестовые и демонстрационные артефакты.
- Закрыть или явно разобрать существующие `TODO`.
- Подготовить проект к более широкому использованию, а не только к локальному или внутреннему.

## Статус по горизонтам

- Быстро 1–4 — выполнено.
- Средне 1–5 — выполнено.
- Долгосрок 1–4 — открыто, отдельный фокус ниже.
- Разбор TODO — закрыт по существенным пунктам, остаточные открытые вопросы перечислены ниже.

## Быстро (выполнено)

### 1. CI и процесс релизов — done

Состояние:

- `.github/workflows/go.yml` запускается на `push` в `main` и на `pull_request`.
- Тесты идут через `go test ./...` и отдельным шагом через `go test -race ./...`.
- CI отвечает только за качество, релизная логика отделена.
- Badge сборки присутствует в README.

### 2. GoReleaser и release workflow — done

Состояние:

- `.goreleaser.yaml` лежит в корне.
- `.github/workflows/release.yml` запускается на `push.tags: ['v*']`, делает checkout с полной историей, setup Go и `goreleaser release --clean`.
- Старый `tag.yml` удалён, автоматического создания тегов нет.
- Версионирование остаётся ручным: подготовил изменения, создал тег, запушил, получил релиз.

### 3. Перенос sample config — done

Состояние:

- Sample-конфиг лежит в `examples/config.json`, в корне его больше нет.
- Все пути обновлены в `config_test.go`, `converter_test.go`, `benchmarks_test.go`, `example_test.go`, `examples/convert_message.go`, `README.md`.
- Константа `CfgJSON` помечена `Deprecated` и теперь указывает на `examples/config.json`. Используется только для обратной совместимости и не упоминается в документации как рекомендуемый путь.

### 4. Уборка шума в workflow и метаданных — done

Состояние:

- `tag.yml` удалён.
- В корне нет временных и демонстрационных артефактов.
- Файл `config.sсhema.json.go` (с кириллической `с` в имени) переименован в `config.schema.json.go`.

## Средне (выполнено)

### 1. README — done

Состояние:

- README перестроен вокруг блоков: что это, какую проблему решает, когда стоит использовать, когда нет, quick start, input/output, mapping-конфиг, обзор архитектуры, ограничения, тесты и benchmark'и.
- Старые и слабые формулировки удалены или переписаны.

### 2. Схемы в README — done

Состояние:

- В README присутствуют Mermaid-схемы потока конвертации, структуры конфига и сравнения двух режимов работы (input-driven и position-driven).

### 3. Конкурентность — done

Состояние:

- В `Converter` нет общего изменяемого состояния. Состояние одного вызова инкапсулировано в `convertCallState`.
- В тестах включён `t.Parallel()` там, где это осмысленно.
- `TestConvertConcurrentSharedConverter` запускает 32 параллельных `Convert` поверх одного `Converter` и сравнивает результаты, фиксируя goroutine safety.
- В CI работает `go test -race ./...`.

### 4. Валидация конфига и обработка options — done

Состояние:

- `Modification.Validate()` проверяет конфликты разделителей, корректность `fields_number`, синтаксис шаблонов и ссылок, ссылочную целостность positions → tags.
- `validateConversionPair` проверяет валидность пары input/output: что `linked` существует и что ссылки `<TAG-N>` указывают на существующие теги и поля во входной модификации.
- Обработка options перенесена в `optionRegistry` (`internal_helpers.go`) с явным `apply` per option и автоматическим сообщением об ошибке вида `available autofill: ...`.

Не закрыто:

- В README пока нет короткого раздела с перечислением поддерживаемых options и их семантикой. Кандидат на следующий маленький патч.

### 5. Поведенческие тесты — done

Состояние:

- В `result_test.go` есть тесты на `Field.Components()`, `Field.Array()`, кэширование и сброс после `ChangeValue`.
- В `converter_test.go` присутствуют сценарии для input-driven и positional conversion, для извлечения aliases (`TestConvertWithUsingAliasesCanReuseSameConverter`), для concurrent reuse converter и fuzz через `FuzzConvert`.
- Комментарии-TODO в тестах либо превращены в тесты, либо удалены.

## Долгосрок

### 1. Улучшить упаковку проекта как продукта

Текущее состояние:

- README уже подаёт проект как config-driven mapping engine, но раздела сравнения с альтернативами нет.

Что сделать:

- Уточнить публичное позиционирование: config-driven ASTM/HL7 lab mapping engine for Go, подходит для LIS bridges, ETL adapters, device integrations.
- Добавить в README раздел сравнения:
  - где проект находится относительно Mirth/NextGen Connect.
  - чем отличается от HAPI HL7v2.
  - чем отличается от шаблонных FHIR-конвертеров.

Ожидаемый результат:

- Посетитель поймёт не только что делает проект, но и почему он вообще существует.

### 2. Добавить CLI

Почему важно:

- CLI — самый прямой способ сделать проект легко пробуемым.
- Многим потенциальным пользователям не захочется писать Go-код ради первой проверки.

Что сделать:

- Добавить CLI в `cmd/hl7conv` или аналогичный каталог.
- Начальный набор команд:
  - `convert`
  - `identify`
  - возможно `validate-config`
- Подключить выпуск CLI-артефактов к GoReleaser (отдельный `builds:` блок в `.goreleaser.yaml`).

Ожидаемый результат:

- Порог входа в проект резко снизится.
- GitHub Releases начнут приносить реальную пользу.

### 3. Доверительные сигналы репозитория

Что сделать:

- Добавить `CONTRIBUTING.md`.
- Добавить issue templates (bug, feature, question) в `.github/ISSUE_TEMPLATE/`.
- Определить стратегию changelog (например, на основе release notes GoReleaser).
- Добавить GitHub topics через UI/API.
- Добавить раздел roadmap в README или отдельный `ROADMAP.md`.
- Принять формат release notes (Keep a Changelog или Conventional Commits + automated section).

Ожидаемый результат:

- Репозиторий выглядит как живой и поддерживаемый open source проект.

### 4. Cleanup публичного API для следующего major

Текущее состояние:

- В публичном API есть legacy-опечатки и неудачные имена.

Кандидаты на cleanup (только следующий major, без breaking change в текущем):

- `IndetifyMsg` → `IdentifyMsg`.
- `TempalateParse` → `TemplateParse`, поле `Tag.Tempalate` → `Tag.Template`.
- Согласовать формулировки и имена ошибок (`unsuccesful` → `unsuccessful`, `tagSturcture` → `tagStructure`, и т.п.).
- Убрать `CfgJSON` или превратить в файл-функцию, не строковый путь.
- Согласовать `(*Converter, error)` сигнатуры — либо реально валидировать в `NewConverter`, либо убрать ошибку.

Ожидаемый результат:

- Публичный API станет аккуратнее без преждевременного breaking change.

## Разбор TODO

### Закрыто

- `config.go` — `TODO: validate` на `ComponentSeparator`: заменено нормальным комментарием. Валидация делается в `Modification.Validate()`.
- `config.go` — `TODO: ADD ALIASES TO MINDTAY_HBL`: устранено, aliases присутствуют в sample config.
- `converter.go` — `TODO: move errors to origin place`: ошибки и конструкторы перенесены в `errors.go`.
- `converter.go` — `todo: clarify ErrOutputTagNotFound`: формулировка ошибки переписана, TODO удалено.
- `converter.go` — `TODO: add opportunity of parallel using converter`: устранено через `convertCallState`, race-тест добавлен.
- `converter.go` — `[TODO] UPGRADE OPTIONS`: заменено на `optionRegistry`-подход.
- `converter.go` — `TODO: check resetPointerIndx`: исчезло вместе с общим состоянием.
- `converter_test.go` — `t.Parallel()`: включён, добавлен concurrent shared converter test.

### Открытые вопросы

- `result.go` — `TODO: check mem and cpu usage` на `otto.New()` в `NewResult`. Действие: сначала бенчмарки, потом ленивая инициализация или пул, без слепых правок.
- `result.go` — `TODO: Does it have correct behaviour?`. Действие: зафиксировать ожидаемое поведение тестом и убрать вопрос.
- `result.go` — `TODO(api): is InsertRow worth adding?` и `TODO(api): is RemoveRowByIndex worth adding?`. Действие: реализовывать только если появится задокументированный пользовательский сценарий, иначе удалить.

## Рекомендуемый порядок выполнения дальше

1. Документировать в README поддерживаемые tag options и их семантику (закрывает «не закрытое» по Средне 4).
2. Бенчмарк `Result.otto` и решение по ленивой инициализации (закрывает первый открытый вопрос по `result.go`).
3. Добавить `CONTRIBUTING.md` и issue templates.
4. Спланировать и реализовать CLI в `cmd/hl7conv` с подключением к GoReleaser.
5. Добавить раздел сравнения с альтернативами в README.
6. Подготовить отдельный план breaking-cleanup для следующего major release.

## Критерии завершения этого плана

Базовая часть плана уже выполнена. Полностью закрытым план будет считаться, когда:

- В README задокументированы tag options и есть раздел сравнения с альтернативами.
- Поставляется CLI-артефакт через GoReleaser.
- В репозитории есть `CONTRIBUTING.md`, issue templates и явная стратегия changelog/release notes.
- Открытые TODO в `result.go` либо закрыты, либо обоснованно отклонены.
- Подготовлен отдельный план breaking-cleanup публичного API для следующего major.
