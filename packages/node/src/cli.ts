#!/usr/bin/env node

import { validate } from "./validator";

async function main(): Promise<void> {
  const args = process.argv.slice(2);

  // Parse simple args and pass through to the binary
  const options: Record<string, string | boolean> = {};
  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === "--schema" || arg === "-s") {
      options.schemaPath = args[++i];
    } else if (arg === "--env" || arg === "-e") {
      options.envPath = args[++i];
    } else if (arg === "--strict") {
      options.strict = true;
    } else if (arg === "--format" || arg === "-f") {
      options.format = args[++i];
    }
  }

  try {
    const result = await validate({
      schemaPath: options.schemaPath as string,
      envPath: options.envPath as string,
      strict: options.strict as boolean,
    });

    if (options.format === "json") {
      console.log(JSON.stringify(result, null, 2));
    } else {
      if (result.valid && result.warnings.length === 0) {
        console.log("✓ All environment variables validated.");
      } else {
        if (!result.valid) {
          console.log(`✗ Environment validation failed (${result.errors.length} error(s))\n`);
          for (const err of result.errors) {
            console.log(`  • ${err.key}`);
            console.log(`    └─ ${err.rule}: ${err.message}`);
          }
        }
        if (result.warnings.length > 0) {
          if (!result.valid) console.log();
          console.log(`⚠ Warnings (${result.warnings.length}):\n`);
          for (const warn of result.warnings) {
            console.log(`  • ${warn.key}`);
            console.log(`    └─ ${warn.rule}: ${warn.message}`);
          }
        }
      }
    }

    process.exit(result.valid ? 0 : 1);
  } catch (err: any) {
    console.error(err.message);
    process.exit(2);
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(2);
});
