class Agentops < Formula
  desc "Local-first observability for AI coding agents"
  homepage "https://github.com/gog1withme/AgentOps"
  url "https://github.com/gog1withme/AgentOps/releases/download/v1.0.0/agentops_1.0.0_darwin_amd64.tar.gz"
  sha256 "PLACEHOLDER_SHA256"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/gog1withme/AgentOps/releases/download/v1.0.0/agentops_1.0.0_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_ARM64"
    end
  end

  def install
    bin.install "agentops"
    (share/"agentops").install "dashboard"
  end

  test do
    assert_match "1.0.0", shell_output("#{bin}/agentops version")
  end
end
