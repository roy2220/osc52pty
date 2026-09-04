# Releasing

Pushing a tag named like `v0.2.1` runs the **Release prebuilt binaries** GitHub
Actions workflow. It creates a GitHub release containing:

- `darwin-amd64` and `darwin-arm64` tarballs;
- the `osc52pty` executable, `LICENSE`, and `README.md` in each tarball;
- SHA-256 hashes in `checksums.txt`.

The workflow can also publish an existing tag. Open it in the Actions tab,
choose **Run workflow**, and enter the tag (for example, `v0.2.0`). If a GitHub
release already exists, its binary assets and checksum file are replaced.

The binaries are currently neither code-signed nor notarized.
