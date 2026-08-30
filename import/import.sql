-- Одноразовый импорт выгрузок easyoffer в схему ezoffer.
--
-- Запуск из корня репозитория (пути в \copy относительные):
--     psql -d ezoffer -v ON_ERROR_STOP=1 -f import/import.sql
-- или через import/run.sh
--
-- Скрипт идемпотентен: повторный запуск обновляет существующие строки.
-- Ключи сопоставления: "sourceId" (вопросы, задачи, приватные собесы),
-- "slug" (тестовые задания, публичные собесы — внешнего id в выгрузке нет).

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------- staging
-- Всё грузим как text и приводим типы при трансформации: сырые CSV чище от
-- этого не станут, а падение на COPY не даст показать внятную ошибку.

CREATE TEMP TABLE "stgQuestions" (
	frequency text, id text, title text, slug text
) ON COMMIT DROP;

CREATE TEMP TABLE "stgQuestionsGrades" (
	grade text, frequency text, id text, title text, slug text
) ON COMMIT DROP;

CREATE TEMP TABLE "stgTasks" (
	rank text, frequency text, id text, slug text, title text, grades text,
	companies text, type text, last_date text, task_url text,
	solutions_url text, content text
) ON COMMIT DROP;

CREATE TEMP TABLE "stgTestTasks" (
	rank text, slug text, title text, date text, grades text, companies text,
	professions_skills text, task_url text, solutions_url text, content text
) ON COMMIT DROP;

CREATE TEMP TABLE "stgPrivateInterviews" (
	id text, title text, slug text, date text, grades text,
	interview_types text, company text, source_url text,
	is_real_interview text, is_visible text, video_is_embeddable text
) ON COMMIT DROP;

CREATE TEMP TABLE "stgPublicInterviews" (
	title text, company text, date text, interview_page text,
	source_url text, slug text
) ON COMMIT DROP;

\copy "stgQuestions" FROM 'data/golang_all_technical_actual.csv' WITH (FORMAT csv, HEADER true)
\copy "stgQuestionsGrades" FROM 'data/golang_all_technical_junior_actual.csv' WITH (FORMAT csv, HEADER true)
\copy "stgQuestionsGrades" FROM 'data/golang_all_technical_middle_actual.csv' WITH (FORMAT csv, HEADER true)
\copy "stgQuestionsGrades" FROM 'data/golang_all_technical_senior_actual.csv' WITH (FORMAT csv, HEADER true)
\copy "stgTasks" FROM 'data/golang_live_code_tasks_actual.csv' WITH (FORMAT csv, HEADER true)
\copy "stgTestTasks" FROM 'data/golang_test_tasks_actual.csv' WITH (FORMAT csv, HEADER true)
\copy "stgPrivateInterviews" FROM 'data/golang_private_interviews.csv' WITH (FORMAT csv, HEADER true)
\copy "stgPublicInterviews" FROM 'data/golang_public_interviews.csv' WITH (FORMAT csv, HEADER true)

-- --------------------------------------------------------- валидация словарей
-- Значения грейдов и типов маппятся в словари схемы. Если в новой выгрузке
-- появится незнакомое значение, импорт должен упасть, а не молча положить NULL.

DO $$
DECLARE bad text;
BEGIN
	SELECT string_agg(DISTINCT g, ', ') INTO bad
	FROM (
		SELECT btrim(unnest(string_to_array(grades, '|'))) AS g FROM "stgTasks"
		UNION ALL SELECT btrim(unnest(string_to_array(grades, '|'))) FROM "stgTestTasks"
		UNION ALL SELECT btrim(unnest(string_to_array(grades, '|'))) FROM "stgPrivateInterviews"
		UNION ALL SELECT btrim(grade) FROM "stgQuestionsGrades"
	) s
	WHERE g <> '' AND g NOT IN ('Junior', 'Middle', 'Senior', 'Lead');
	IF bad IS NOT NULL THEN
		RAISE EXCEPTION 'неизвестный грейд в выгрузке: %', bad;
	END IF;

	SELECT string_agg(DISTINCT btrim(type), ', ') INTO bad
	FROM "stgTasks"
	WHERE btrim(coalesce(type, '')) <> ''
	  AND btrim(type) NOT IN ('Live-coding', 'Алгоритмы', 'System Design');
	IF bad IS NOT NULL THEN
		RAISE EXCEPTION 'неизвестный тип задачи: %', bad;
	END IF;

	SELECT string_agg(DISTINCT t, ', ') INTO bad
	FROM (
		SELECT btrim(unnest(string_to_array(interview_types, '|'))) AS t
		FROM "stgPrivateInterviews"
	) s
	WHERE t <> ''
	  AND t NOT IN ('Техническое', 'Live-coding', 'Алгоритмическое',
	                'HR-скрининг', 'Финальное', 'System Design');
	IF bad IS NOT NULL THEN
		RAISE EXCEPTION 'неизвестный тип собеседования: %', bad;
	END IF;
END $$;

-- ------------------------------------------------------------- справочники
-- Компании приходят из четырёх источников с одинаковым написанием, поэтому
-- склеиваем по имени. Если источники начнут расходиться («OZON» / «Ozon»),
-- сюда добавится таблица алиасов.

INSERT INTO "companies" ("name")
SELECT DISTINCT btrim(name)
FROM (
	SELECT unnest(string_to_array(companies, '|')) AS name FROM "stgTasks"
	UNION ALL SELECT unnest(string_to_array(companies, '|')) FROM "stgTestTasks"
	UNION ALL SELECT company FROM "stgPrivateInterviews"
	UNION ALL SELECT company FROM "stgPublicInterviews"
) s
WHERE btrim(coalesce(name, '')) <> ''
ON CONFLICT ("name") DO NOTHING;

INSERT INTO "skills" ("name")
SELECT DISTINCT btrim(name)
FROM (
	SELECT unnest(string_to_array(professions_skills, '|')) AS name FROM "stgTestTasks"
) s
WHERE btrim(coalesce(name, '')) <> ''
ON CONFLICT ("name") DO NOTHING;

-- ----------------------------------------------------------------- вопросы
-- В выгрузке «шанс» — доля 0..1, в схеме и в API это проценты 0..100.

INSERT INTO "questions" ("sourceId", "slug", "content", "frequency")
SELECT id::bigint, slug, title, round(frequency::numeric * 100, 2)::real
FROM "stgQuestions"
ON CONFLICT ("sourceId") DO UPDATE SET
	"slug" = EXCLUDED."slug",
	"content" = EXCLUDED."content",
	"frequency" = EXCLUDED."frequency";

-- Один вопрос живёт на нескольких грейдах, и «шанс» на каждом свой.
INSERT INTO "questionsGrades" ("questionId", "grade", "frequency")
SELECT q."questionId", lower(btrim(s.grade)), round(s.frequency::numeric * 100, 2)::real
FROM "stgQuestionsGrades" s
JOIN "questions" q ON q."sourceId" = s.id::bigint
ON CONFLICT ("questionId", "grade") DO UPDATE SET
	"frequency" = EXCLUDED."frequency";

-- ------------------------------------------------------------------ задачи
-- frequency в выгрузке — константа для всех строк, не переносим.
-- Сниппет Go-кода отдельным полем не приходит: он лежит внутри content
-- markdown-блоком, поэтому content сохраняем как есть.

INSERT INTO "tasks" ("sourceId", "slug", "title", "grades", "type", "content",
                     "sourceUrl", "rank", "lastDate")
SELECT
	s.id::bigint,
	s.slug,
	s.title,
	ARRAY(
		SELECT lower(btrim(g))
		FROM unnest(string_to_array(s.grades, '|')) g
		WHERE btrim(g) <> ''
	)::varchar[],
	CASE btrim(coalesce(s.type, ''))
		WHEN 'Live-coding' THEN 'liveCoding'
		WHEN 'Алгоритмы' THEN 'algorithms'
		WHEN 'System Design' THEN 'systemDesign'
	END,
	s.content,
	s.task_url,
	s.rank::int,
	s.last_date::timestamptz
FROM "stgTasks" s
ON CONFLICT ("sourceId") DO UPDATE SET
	"slug" = EXCLUDED."slug",
	"title" = EXCLUDED."title",
	"grades" = EXCLUDED."grades",
	"type" = EXCLUDED."type",
	"content" = EXCLUDED."content",
	"sourceUrl" = EXCLUDED."sourceUrl",
	"rank" = EXCLUDED."rank",
	"lastDate" = EXCLUDED."lastDate";

INSERT INTO "tasksCompanies" ("taskId", "companyId")
SELECT t."taskId", c."companyId"
FROM "stgTasks" s
JOIN "tasks" t ON t."sourceId" = s.id::bigint
CROSS JOIN LATERAL unnest(string_to_array(s.companies, '|')) AS u(name)
JOIN "companies" c ON c."name" = btrim(u.name)
WHERE btrim(u.name) <> ''
ON CONFLICT DO NOTHING;

-- -------------------------------------------------------- тестовые задания
-- Внешнего id в выгрузке нет, ключ сопоставления — slug.

INSERT INTO "testAssignments" ("slug", "title", "grades", "content",
                               "sourceUrl", "publishedDate")
SELECT
	s.slug,
	s.title,
	ARRAY(
		SELECT lower(btrim(g))
		FROM unnest(string_to_array(s.grades, '|')) g
		WHERE btrim(g) <> ''
	)::varchar[],
	s.content,
	s.task_url,
	s.date::date
FROM "stgTestTasks" s
ON CONFLICT ("slug") DO UPDATE SET
	"title" = EXCLUDED."title",
	"grades" = EXCLUDED."grades",
	"content" = EXCLUDED."content",
	"sourceUrl" = EXCLUDED."sourceUrl",
	"publishedDate" = EXCLUDED."publishedDate";

INSERT INTO "testAssignmentsCompanies" ("testAssignmentId", "companyId")
SELECT ta."testAssignmentId", c."companyId"
FROM "stgTestTasks" s
JOIN "testAssignments" ta ON ta."slug" = s.slug
CROSS JOIN LATERAL unnest(string_to_array(s.companies, '|')) AS u(name)
JOIN "companies" c ON c."name" = btrim(u.name)
WHERE btrim(u.name) <> ''
ON CONFLICT DO NOTHING;

INSERT INTO "testAssignmentsSkills" ("testAssignmentId", "skillId")
SELECT ta."testAssignmentId", sk."skillId"
FROM "stgTestTasks" s
JOIN "testAssignments" ta ON ta."slug" = s.slug
CROSS JOIN LATERAL unnest(string_to_array(s.professions_skills, '|')) AS u(name)
JOIN "skills" sk ON sk."name" = btrim(u.name)
WHERE btrim(u.name) <> ''
ON CONFLICT DO NOTHING;

-- ------------------------------------------------- записи собеседований
-- Приватные: видео лежит в Telegram, ссылки на YouTube нет вообще.
-- До выгрузки на VFS "sourceUrl" — единственная ссылка на запись.

INSERT INTO "interviews" ("sourceId", "slug", "title", "grades", "types",
                          "companyId", "sourceUrl", "isEmbeddable",
                          "isReal", "isVisible", "publishedDate")
SELECT
	s.id::bigint,
	s.slug,
	s.title,
	ARRAY(
		SELECT lower(btrim(g))
		FROM unnest(string_to_array(s.grades, '|')) g
		WHERE btrim(g) <> ''
	)::varchar[],
	ARRAY(
		SELECT CASE btrim(t)
			WHEN 'Техническое' THEN 'technical'
			WHEN 'Live-coding' THEN 'liveCoding'
			WHEN 'Алгоритмическое' THEN 'algorithmic'
			WHEN 'HR-скрининг' THEN 'hrScreening'
			WHEN 'Финальное' THEN 'final'
			WHEN 'System Design' THEN 'systemDesign'
		END
		FROM unnest(string_to_array(s.interview_types, '|')) t
		WHERE btrim(t) <> ''
	)::varchar[],
	c."companyId",
	s.source_url,
	s.video_is_embeddable::boolean,
	s.is_real_interview::boolean,
	s.is_visible::boolean,
	s.date::date
FROM "stgPrivateInterviews" s
LEFT JOIN "companies" c ON c."name" = btrim(s.company)
ON CONFLICT ("slug") DO UPDATE SET
	"sourceId" = EXCLUDED."sourceId",
	"title" = EXCLUDED."title",
	"grades" = EXCLUDED."grades",
	"types" = EXCLUDED."types",
	"companyId" = EXCLUDED."companyId",
	"sourceUrl" = EXCLUDED."sourceUrl",
	"isEmbeddable" = EXCLUDED."isEmbeddable",
	"isReal" = EXCLUDED."isReal",
	"isVisible" = EXCLUDED."isVisible",
	"publishedDate" = EXCLUDED."publishedDate";

-- Публичные: грейд и тип в выгрузке не распарсены, компания заполнена у
-- единиц — оставляем пустые массивы и NULL, каталог их всё равно показывает.
-- "isEmbeddable" = true: это ссылки на YouTube, их фронт встраивает плеером.

INSERT INTO "interviews" ("slug", "title", "companyId", "youtubeUrl",
                          "sourceUrl", "isEmbeddable", "publishedDate")
SELECT
	s.slug,
	s.title,
	c."companyId",
	s.source_url,
	s.interview_page,
	true,
	s.date::date
FROM "stgPublicInterviews" s
LEFT JOIN "companies" c ON c."name" = btrim(s.company) AND btrim(s.company) <> ''
ON CONFLICT ("slug") DO UPDATE SET
	"title" = EXCLUDED."title",
	"companyId" = EXCLUDED."companyId",
	"youtubeUrl" = EXCLUDED."youtubeUrl",
	"sourceUrl" = EXCLUDED."sourceUrl",
	"isEmbeddable" = EXCLUDED."isEmbeddable",
	"publishedDate" = EXCLUDED."publishedDate";

COMMIT;

-- ------------------------------------------------------------------ отчёт

SELECT 'questions' AS entity, count(*) FROM "questions"
UNION ALL SELECT 'questionsGrades', count(*) FROM "questionsGrades"
UNION ALL SELECT 'tasks', count(*) FROM "tasks"
UNION ALL SELECT 'tasksCompanies', count(*) FROM "tasksCompanies"
UNION ALL SELECT 'testAssignments', count(*) FROM "testAssignments"
UNION ALL SELECT 'testAssignmentsCompanies', count(*) FROM "testAssignmentsCompanies"
UNION ALL SELECT 'testAssignmentsSkills', count(*) FROM "testAssignmentsSkills"
UNION ALL SELECT 'interviews', count(*) FROM "interviews"
UNION ALL SELECT 'companies', count(*) FROM "companies"
UNION ALL SELECT 'skills', count(*) FROM "skills";
