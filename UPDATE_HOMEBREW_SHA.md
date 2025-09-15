# 🔐 Updating Homebrew Formula SHA256

## After Creating GitHub Release

Once you create the v0.2.0 release on GitHub, you need to update the SHA256 hash in the Homebrew formula:

### 1. Calculate the SHA256

```bash
# Download and calculate SHA256 for the release tarball
curl -L https://github.com/jakeasaurus/lazytodo/archive/refs/tags/v0.2.0.tar.gz | shasum -a 256
```

### 2. Update the Formula

Edit `/path/to/homebrew-tap/Formula/lazytodo.rb` and replace:

```ruby
sha256 "" # This will be calculated when the release is created
```

With:

```ruby
sha256 "THE_CALCULATED_HASH_HERE"
```

### 3. Test the Formula

```bash
# Test installation from your tap
brew install --build-from-source jakeasaurus/tap/lazytodo

# Test that it works
lazytodo --version
```

### 4. Commit and Push

```bash
cd /path/to/homebrew-tap
git add Formula/lazytodo.rb
git commit -m "Update SHA256 for lazytodo v0.2.0"
git push origin main
```

## 🎆 Result

Users can now install with:

```bash
brew tap jakeasaurus/tap
brew install lazytodo
```

---

**Note:** Delete this file after updating the formula! 🗑️