# Documentation: https://docs.brew.sh/Formula-Cookbook
#                https://rubydoc.brew.sh/Formula
# PLEASE REMOVE ALL GENERATED COMMENTS BEFORE SUBMITTING YOUR PULL REQUEST!

class Lazytodo < Formula
  desc "⚡ A synthwave TUI wrapper for todo.txt with electric vibes"
  homepage "https://github.com/zachreborn/lazytodo"
  url "https://github.com/zachreborn/lazytodo/archive/v0.2.0.tar.gz"
  # sha256 will be calculated and updated when releasing
  sha256 ""
  license "MIT"

  depends_on "go" => :build
  depends_on "todo-txt"

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "-o", bin/"lazytodo"
  end

  test do
    # Test that the binary was installed correctly
    assert_match "lazytodo v", shell_output("#{bin}/lazytodo --version")
    
    # Test help output
    assert_match "TUI wrapper for todo.txt", shell_output("#{bin}/lazytodo --help")
  end

  def caveats
    <<~EOS
      🌆 Welcome to the neon future of productivity! ⚡
      
      lazytodo is now installed and ready to use.
      
      🚀 Get started:
        lazytodo
      
      💡 For help:
        lazytodo --help
      
      📁 lazytodo will use your existing todo.txt setup or create
         ~/todo.txt if no configuration is found.
      
      🎭 Powered by Charm - https://charm.sh
    EOS
  end
end