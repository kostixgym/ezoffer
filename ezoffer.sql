CREATE TABLE "users" (
	"userId" BIGSERIAL NOT NULL,
	CONSTRAINT "pk_users" PRIMARY KEY("userId")
);

CREATE TABLE "questions" (
	"questionId" BIGSERIAL NOT NULL,
	"slug" varchar(255) UNIQUE NOT NULL,
	"frequency" real NOT NULL,
	"grades" varchar[] NOT NULL,
	"content" text NOT NULL,
	"createdAt" timestamptz NOT NULL,
	CONSTRAINT "pk_questions" PRIMARY KEY("questionId")
);

CREATE TABLE "tasks" (
	"taskId" BIGSERIAL NOT NULL,
	"slug" varchar(255) UNIQUE NOT NULL,
	"title" varchar(255) NOT NULL,
	"grades" varchar[] NOT NULL,
	"type" varchar(255) NOT NULL,
	"content" text NOT NULL,
	"createdAt" timestamptz NOT NULL,
	CONSTRAINT "pk_tasks" PRIMARY KEY("taskId")
);

CREATE TABLE "testAssignments" (
	"testAssignmentId" BIGSERIAL NOT NULL,
	"slug" varchar(255) UNIQUE NOT NULL,
	"title" varchar(255) NOT NULL,
	"grades" varchar[] NOT NULL,
	"content" text NOT NULL,
	"publishedDate" date NOT NULL,
	"createdAt" timestamptz NOT NULL,
	CONSTRAINT "pk_testAssignments" PRIMARY KEY("testAssignmentId")
);

CREATE TABLE "interviews" (
	"interviewId" BIGSERIAL NOT NULL,
	"slug" varchar(255) UNIQUE NOT NULL,
	"title" varchar(255) NOT NULL,
	"grades" varchar[] NOT NULL,
	"type" varchar[] NOT NULL,
	"videoLink" varchar(255) NOT NULL,
	"isReal" boolean NOT NULL,
	"publishedDate" date NOT NULL,
	"createdAt" timestamptz NOT NULL,
	"companyId" BIGINT NOT NULL,
	CONSTRAINT "pk_interviews" PRIMARY KEY("interviewId")
);

CREATE TABLE "companies" (
	"companyId" BIGSERIAL NOT NULL,
	"name" varchar(255) UNIQUE NOT NULL,
	CONSTRAINT "pk_companies" PRIMARY KEY("companyId")
);

CREATE TABLE "skills" (
	"skillId" BIGSERIAL NOT NULL,
	"name" varchar(255) NOT NULL,
	CONSTRAINT "pk_skills" PRIMARY KEY("skillId")
);

CREATE TABLE "tasksCompanies" (
	"taskId" BIGINT NOT NULL,
	"companyId" BIGINT NOT NULL,
	CONSTRAINT "pk_tasksCompanies" PRIMARY KEY("taskId", "companyId")
);

CREATE TABLE "testAssignmentsCompanies" (
	"testAssignmentId" BIGINT NOT NULL,
	"companyId" BIGINT NOT NULL,
	CONSTRAINT "pk_testAssignmentsCompanies" PRIMARY KEY("testAssignmentId", "companyId")
);

CREATE TABLE "testAssignmentsSkills" (
	"testAssignmentId" BIGINT NOT NULL,
	"skillId" BIGINT NOT NULL,
	CONSTRAINT "pk_testAssignmentsSkills" PRIMARY KEY("testAssignmentId", "skillId")
);

CREATE TABLE "comments" (
	"commentId" BIGSERIAL NOT NULL,
	"likesCnt" BIGINT NOT NULL,
	"content" text NOT NULL,
	"interviewId" BIGINT,
	"questionId" BIGINT,
	"testAssignmentId" BIGINT,
	"taskId" BIGINT,
	CONSTRAINT "pk_comments" PRIMARY KEY("commentId"),
	CONSTRAINT "chk_one_entity" CHECK (
        ("interviewId" IS NOT NULL)::int + ("questionId" IS NOT NULL)::int + 
        ("testAssignmentId" IS NOT NULL)::int + ("taskId" IS NOT NULL)::int = 1
    )
);

CREATE INDEX "ix_interviews_companyId" ON "interviews" (
	"companyId"
);

CREATE INDEX "ix_tasksCompanies_taskId" ON "tasksCompanies" (
	"taskId"
);

CREATE INDEX "ix_tasksCompanies_companyId" ON "tasksCompanies" (
	"companyId"
);

CREATE INDEX "ix_testAssignmentsCompanies_testAssignmentId" ON "testAssignmentsCompanies" (
	"testAssignmentId"
);

CREATE INDEX "ix_testAssignmentsCompanies_companyId" ON "testAssignmentsCompanies" (
	"companyId"
);

CREATE INDEX "ix_testAssignmentsSkills_testAssignmentId" ON "testAssignmentsSkills" (
	"testAssignmentId"
);

CREATE INDEX "ix_testAssignmentsSkills_skillId" ON "testAssignmentsSkills" (
	"skillId"
);

CREATE INDEX "ix_comments_interviewId" ON "comments" (
	"interviewId"
);

CREATE INDEX "ix_comments_questionId" ON "comments" (
	"questionId"
);

CREATE INDEX "ix_comments_testAssignmentId" ON "comments" (
	"testAssignmentId"
);

CREATE INDEX "ix_comments_taskId" ON "comments" (
	"taskId"
);

ALTER TABLE "interviews" ADD CONSTRAINT "fk_interviews_companyId" FOREIGN KEY ("companyId")
	REFERENCES "companies"("companyId")
	ON DELETE RESTRICT
	ON UPDATE RESTRICT
	NOT DEFERRABLE;

ALTER TABLE "tasksCompanies" ADD CONSTRAINT "fk_tasksCompanies_taskId" FOREIGN KEY ("taskId")
	REFERENCES "tasks"("taskId")
	ON DELETE RESTRICT
	ON UPDATE RESTRICT
	NOT DEFERRABLE;

ALTER TABLE "tasksCompanies" ADD CONSTRAINT "fk_tasksCompanies_companyId" FOREIGN KEY ("companyId")
	REFERENCES "companies"("companyId")
	ON DELETE RESTRICT
	ON UPDATE RESTRICT
	NOT DEFERRABLE;

ALTER TABLE "testAssignmentsCompanies" ADD CONSTRAINT "fk_testAssignmentsCompanies_testAssignmentId" FOREIGN KEY ("testAssignmentId")
	REFERENCES "testAssignments"("testAssignmentId")
	ON DELETE RESTRICT
	ON UPDATE RESTRICT
	NOT DEFERRABLE;

ALTER TABLE "testAssignmentsCompanies" ADD CONSTRAINT "fk_testAssignmentsCompanies_companyId" FOREIGN KEY ("companyId")
	REFERENCES "companies"("companyId")
	ON DELETE RESTRICT
	ON UPDATE RESTRICT
	NOT DEFERRABLE;

ALTER TABLE "testAssignmentsSkills" ADD CONSTRAINT "fk_testAssignmentsSkills_testAssignmentId" FOREIGN KEY ("testAssignmentId")
	REFERENCES "testAssignments"("testAssignmentId")
	ON DELETE RESTRICT
	ON UPDATE RESTRICT
	NOT DEFERRABLE;

ALTER TABLE "testAssignmentsSkills" ADD CONSTRAINT "fk_testAssignmentsSkills_skillId" FOREIGN KEY ("skillId")
	REFERENCES "skills"("skillId")
	ON DELETE RESTRICT
	ON UPDATE RESTRICT
	NOT DEFERRABLE;

ALTER TABLE "comments" ADD CONSTRAINT "fk_comments_interviewId" FOREIGN KEY ("interviewId")
	REFERENCES "interviews"("interviewId")
	ON DELETE RESTRICT
	ON UPDATE RESTRICT
	NOT DEFERRABLE;

ALTER TABLE "comments" ADD CONSTRAINT "fk_comments_questionId" FOREIGN KEY ("questionId")
	REFERENCES "questions"("questionId")
	ON DELETE RESTRICT
	ON UPDATE RESTRICT
	NOT DEFERRABLE;

ALTER TABLE "comments" ADD CONSTRAINT "fk_comments_testAssignmentId" FOREIGN KEY ("testAssignmentId")
	REFERENCES "testAssignments"("testAssignmentId")
	ON DELETE RESTRICT
	ON UPDATE RESTRICT
	NOT DEFERRABLE;

ALTER TABLE "comments" ADD CONSTRAINT "fk_comments_taskId" FOREIGN KEY ("taskId")
	REFERENCES "tasks"("taskId")
	ON DELETE RESTRICT
	ON UPDATE RESTRICT
	NOT DEFERRABLE;

