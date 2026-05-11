import * as fs from "fs";
import * as https from "https";
import * as os from "os";
import * as path from "path";

const VERSION = "0.1.8";
const REPO = "firasmosbehi/envguard";

function getPlatform(): string {
  const platform = os.platform();
  const arch = os.arch();

  const platformMap: Record<string, string> = {
    darwin: "darwin",
    linux: "linux",
    win32: "windows",
  };

  const archMap: Record<string, string> = {
    x64: "amd64",
    arm64: "arm64",
  };

  const p = platformMap[platform];
  const a = archMap[arch];

  if (!p || !a) {
    throw new Error(`Unsupported platform: ${platform}/${arch}`);
  }

  if (p === "darwin" && a === "amd64" && process.env.APPLE_SILICON_ROSETTA) {
    // Allow override for testing
  }

  return `${p}-${a}`;
}

function getBinaryName(): string {
  return os.platform() === "win32" ? "envguard.exe" : "envguard";
}

function getBinaryDir(): string {
  // Store binary inside the package's dist folder
  return path.join(__dirname);
}

function downloadFile(url: string, dest: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https
      .get(url, { headers: { "User-Agent": "envguard-node-installer" } }, (response) => {
        if (response.statusCode === 302 || response.statusCode === 301) {
          const redirectUrl = response.headers.location;
          if (!redirectUrl) {
            reject(new Error("Redirect without location header"));
            return;
          }
          downloadFile(redirectUrl, dest).then(resolve).catch(reject);
          return;
        }

        if (response.statusCode !== 200) {
          reject(new Error(`Download failed with status ${response.statusCode}`));
          return;
        }

        response.pipe(file);
        file.on("finish", () => {
          file.close();
          resolve();
        });
      })
      .on("error", (err) => {
        fs.unlink(dest, () => {});
        reject(err);
      });
  });
}

async function main(): Promise<void> {
  const platform = getPlatform();
  const binaryName = getBinaryName();
  const binaryDir = getBinaryDir();
  const binaryPath = path.join(binaryDir, binaryName);

  // Check if binary already exists
  if (fs.existsSync(binaryPath)) {
    console.log(`EnvGuard binary already exists at ${binaryPath}`);
    return;
  }

  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/envguard-${platform}${
    os.platform() === "win32" ? ".exe" : ""
  }`;

  console.log(`Downloading EnvGuard ${VERSION} for ${platform}...`);
  console.log(`URL: ${url}`);

  try {
    await downloadFile(url, binaryPath);
    fs.chmodSync(binaryPath, 0o755);
    console.log(`EnvGuard binary installed at ${binaryPath}`);
  } catch (err) {
    console.error(`Failed to download EnvGuard binary: ${err}`);
    console.error("You can manually download it from:");
    console.error(`https://github.com/${REPO}/releases/tag/v${VERSION}`);
    process.exit(1);
  }
}

// Run if called directly (postinstall)
if (require.main === module) {
  main().catch((err) => {
    console.error(err);
    process.exit(1);
  });
}

export { getBinaryDir, getBinaryName, getPlatform };
