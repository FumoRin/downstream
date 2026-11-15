import test from "node:test";
import assert from "node:assert";
import { formatETA } from "../src/utils/formatters.ts";

test("formatETA - invalid or non-positive inputs", () => {
  assert.strictEqual(formatETA(0), "--");
  assert.strictEqual(formatETA(-10), "--");
  assert.strictEqual(formatETA(NaN), "--");
  assert.strictEqual(formatETA(Infinity), "--");
  // @ts-expect-error test runtime boundary
  assert.strictEqual(formatETA(null), "--");
  // @ts-expect-error test runtime boundary
  assert.strictEqual(formatETA(undefined), "--");
});

test("formatETA - adaptive mode (default)", () => {
  assert.strictEqual(formatETA(45), "00:45");
  assert.strictEqual(formatETA(125), "02:05");
  assert.strictEqual(formatETA(3599), "59:59");
  assert.strictEqual(formatETA(3600), "01:00:00");
  assert.strictEqual(formatETA(3665), "01:01:05");
  assert.strictEqual(formatETA(7325), "02:02:05");
  assert.strictEqual(formatETA(360000), "100:00:00");
});

test("formatETA - fixed mode", () => {
  assert.strictEqual(formatETA(45, "fixed"), "00:00:45");
  assert.strictEqual(formatETA(125, "fixed"), "00:02:05");
  assert.strictEqual(formatETA(3599, "fixed"), "00:59:59");
  assert.strictEqual(formatETA(3600, "fixed"), "01:00:00");
  assert.strictEqual(formatETA(3665, "fixed"), "01:01:05");
});
