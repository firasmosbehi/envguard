import { describe, it } from "node:test";
import assert from "node:assert";
import { getBinaryName, getBinaryDir, getPlatform } from "../install";
import * as path from "path";

describe("install", () => {
  describe("getBinaryName", () => {
    it("should return envguard.exe on Windows", () => {
      const originalPlatform = Object.getOwnPropertyDescriptor(process, "platform");
      Object.defineProperty(process, "platform", {
        value: "win32",
      });
      try {
        assert.strictEqual(getBinaryName(), "envguard.exe");
      } finally {
        if (originalPlatform) {
          Object.defineProperty(process, "platform", originalPlatform);
        }
      }
    });

    it("should return envguard on non-Windows", () => {
      const originalPlatform = Object.getOwnPropertyDescriptor(process, "platform");
      Object.defineProperty(process, "platform", {
        value: "linux",
      });
      try {
        assert.strictEqual(getBinaryName(), "envguard");
      } finally {
        if (originalPlatform) {
          Object.defineProperty(process, "platform", originalPlatform);
        }
      }
    });
  });

  describe("getBinaryDir", () => {
    it("should return a directory path", () => {
      const dir = getBinaryDir();
      assert.strictEqual(typeof dir, "string");
      assert.ok(path.isAbsolute(dir) || dir.startsWith("."));
    });
  });

  describe("getPlatform", () => {
    it("should return linux-amd64 for Linux x64", () => {
      const originalPlatform = Object.getOwnPropertyDescriptor(process, "platform");
      const originalArch = Object.getOwnPropertyDescriptor(process, "arch");
      Object.defineProperty(process, "platform", { value: "linux" });
      Object.defineProperty(process, "arch", { value: "x64" });
      try {
        assert.strictEqual(getPlatform(), "linux-amd64");
      } finally {
        if (originalPlatform) {
          Object.defineProperty(process, "platform", originalPlatform);
        }
        if (originalArch) {
          Object.defineProperty(process, "arch", originalArch);
        }
      }
    });

    it("should return darwin-arm64 for macOS arm64", () => {
      const originalPlatform = Object.getOwnPropertyDescriptor(process, "platform");
      const originalArch = Object.getOwnPropertyDescriptor(process, "arch");
      Object.defineProperty(process, "platform", { value: "darwin" });
      Object.defineProperty(process, "arch", { value: "arm64" });
      try {
        assert.strictEqual(getPlatform(), "darwin-arm64");
      } finally {
        if (originalPlatform) {
          Object.defineProperty(process, "platform", originalPlatform);
        }
        if (originalArch) {
          Object.defineProperty(process, "arch", originalArch);
        }
      }
    });

    it("should throw for unsupported platform", () => {
      const originalPlatform = Object.getOwnPropertyDescriptor(process, "platform");
      const originalArch = Object.getOwnPropertyDescriptor(process, "arch");
      Object.defineProperty(process, "platform", { value: "freebsd" });
      Object.defineProperty(process, "arch", { value: "x64" });
      try {
        assert.throws(() => getPlatform(), /Unsupported platform/);
      } finally {
        if (originalPlatform) {
          Object.defineProperty(process, "platform", originalPlatform);
        }
        if (originalArch) {
          Object.defineProperty(process, "arch", originalArch);
        }
      }
    });

    it("should throw for unsupported arch", () => {
      const originalPlatform = Object.getOwnPropertyDescriptor(process, "platform");
      const originalArch = Object.getOwnPropertyDescriptor(process, "arch");
      Object.defineProperty(process, "platform", { value: "linux" });
      Object.defineProperty(process, "arch", { value: "mips" });
      try {
        assert.throws(() => getPlatform(), /Unsupported platform/);
      } finally {
        if (originalPlatform) {
          Object.defineProperty(process, "platform", originalPlatform);
        }
        if (originalArch) {
          Object.defineProperty(process, "arch", originalArch);
        }
      }
    });
  });
});
