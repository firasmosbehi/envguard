class Envguard < Formula
  desc "Validate .env files against schemas"
  homepage "https://github.com/firasmosbehi/envguard"
  version "0.1.7"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/firasmosbehi/envguard/releases/download/v0.1.7/envguard-darwin-arm64"
      sha256 :no_check
    else
      url "https://github.com/firasmosbehi/envguard/releases/download/v0.1.7/envguard-darwin-amd64"
      sha256 :no_check
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/firasmosbehi/envguard/releases/download/v0.1.7/envguard-linux-arm64"
      sha256 :no_check
    else
      url "https://github.com/firasmosbehi/envguard/releases/download/v0.1.7/envguard-linux-amd64"
      sha256 :no_check
    end
  end

  def install
    bin.install Dir["envguard-*"].first => "envguard"
  end

  test do
    (testpath/"envguard.yaml").write <<~YAML
      version: "1.0"
      env:
        TEST:
          type: string
          required: true
    YAML
    (testpath/".env").write "TEST=hello\n"
    assert_match "All environment variables validated", shell_output("#{bin}/envguard validate --schema envguard.yaml --env .env")
  end
end
