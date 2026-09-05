import { describe, it, expect } from "vitest";
import { DAILY_REVIEW_PROMPTS, emptyAnswers, fetchDailyReview, saveDailyReview } from "./reviewData";

describe("emptyAnswers", () => {
  it("has one blank entry per fixed prompt", () => {
    const answers = emptyAnswers();
    expect(Object.keys(answers).sort()).toEqual(DAILY_REVIEW_PROMPTS.map((p) => p.key).sort());
    expect(Object.values(answers).every((v) => v === "")).toBe(true);
  });
});

describe("fetchDailyReview / saveDailyReview", () => {
  it("returns null for a date with no saved review", async () => {
    expect(await fetchDailyReview("2020-01-01")).toBeNull();
  });

  it("round-trips a saved review", async () => {
    const answers = { ...emptyAnswers(), wentWell: "Shipped Phase 10" };
    const saved = await saveDailyReview("2026-08-15", answers);
    expect(saved.date).toBe("2026-08-15");
    expect(saved.answers.wentWell).toBe("Shipped Phase 10");

    const fetched = await fetchDailyReview("2026-08-15");
    expect(fetched?.answers.wentWell).toBe("Shipped Phase 10");
  });

  it("overwrites a previous save for the same date", async () => {
    await saveDailyReview("2026-08-16", { ...emptyAnswers(), grateful: "coffee" });
    await saveDailyReview("2026-08-16", { ...emptyAnswers(), grateful: "tea" });
    const fetched = await fetchDailyReview("2026-08-16");
    expect(fetched?.answers.grateful).toBe("tea");
  });
});
