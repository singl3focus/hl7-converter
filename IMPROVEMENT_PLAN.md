# План развития hl7-converter

Обновлено 28.10.2026. База: v2.6.0, main 9c438ad. Изменения подготовлены в ветке fix/mapping-reliability. Локальная реализация не означает публикацию релиза.

## Реализовано

- CI/unit/race, CLI, CONTRIBUTING, issue templates и GoReleaser уже существовали в v2.6.0.
- Options/autofill документированы, GitHub topics выставлены. Устаревшие незакрытые пункты удалены.
- ParseInput/input order conversion проверяют scanner.Err и не возвращают успешный неполный результат.
- Field/component разбираются как целые числа: .10/.12 не превращаются в .1.
- IdentifyMsg не меняет общий Types. Пересекающиеся типы дают явную ambiguity error.
- Constructor проверяет nil params/options, модификации, связи и output counts. fields_number=-1 сохраняется для input/unused tags.
- positions принимают канонические положительные ключи; многосимвольный separator работает в обоих режимах.
- Error.Unwrap и sentinel identity заменили сравнение текста ошибок.
- TemplateParse работает с байтами и не падает на Unicode; вложенные ссылки отвергаются.
- IndetifyMsg/TempalateParse сохранены, правильные имена добавлены совместимо.
- CLI: справка, ошибки записи, -raw, семантическая валидация всех блоков и выбранной пары.
- check: version-1 fixtures с exact bytes, aliases, expected error code, inline/file input/output.
- explain: статические связи output fields → input links/defaults; JSON diagnostics без payload.
- JSON Schema встроена из одного файла через go:embed.
- Otto создаётся при первом UseScript и сохраняет script state.
- FuzzConvert проверяет recovered panic и result.String/Bytes; добавлен fuzz TemplateParse.
- Синтетические examples/mappings выполняются в тестах и CI; README/GoDoc описывают numbering/repeats/limits/stdout.
- CI разделяет compatibility Go 1.22.3 и поддерживаемую 1.26.x. Release использует поддерживаемый компилятор и проверки до GoReleaser.

## Перед публикацией

- Локальные проверки Go 1.22.3/1.26.10, race/vet, три fuzz-набора и security выполнены. Покрытие: library 79.6%, CLI 78.5%.
- Review выполнен локально; независимый review перед merge остаётся полезным.
- Windows CLI smoke и сборка всех шести targets выполнены. Кросс-компиляция не заменяет запуск на Linux/macOS.
- Прогнать GitHub CI на ветке/PR; локальные проверки не подтверждают GitHub Actions.
- Проверить GoReleaser packaging и build info свежего артефакта.
- Выпустить новый minor: добавлены команды/public API. Не перезаписывать v2.6.0.
- Новый workflow исправляет будущие сборки, но не уже скачанные Go 1.22.3 бинарники.

## Следующие задачи по подтверждённым сценариям

- Отдельные block/tag/field координаты в JSON diagnostics.
- Настраиваемый предел строки при реальном input >64 KiB; сейчас явный отказ.
- Независимые fixtures и спецификация для grouping/repeats/escaping до расширения parser layer.
- Обновление транзитивных зависимостей отдельным изменением с проверкой minimum Go. Module-only warnings не равны доказанной достижимости.
- Изменения absent linked rows/cross-tag repeats требуют совместимого дизайна.
- Проверять востребованность check/explain по задачам интеграторов без обещаний роста stars/downloads.

## Отложено

- InsertRow/RemoveRowByIndex до конкретного caller.
- Полный integration server, MLLP daemon, queues/persistence/retries.
- FHIR и полная HL7/ASTM conformance без отдельного профиля.
- Переименование Tag.Tempalate/удаление legacy API до отдельного compatibility/major плана.
- Общий VM pool: достаточно ленивой инициализации для измеренной проблемы.
- Отдельный сайт: сейчас развиваются библиотека, CLI и GitHub.
