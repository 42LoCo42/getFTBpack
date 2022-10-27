# getFTBpack
How to run:
- Get the pack and version ID from the FTB website
([example](https://www.feed-the-beast.com/modpacks/99-ftb-inferno),
has pack ID 99 and the latest release currently has version ID 2270)
- Run: `go run . <pack ID> <version ID>` (example: `go run . 99 2270`)
- Wait for completion
- All target versions (Minecraft, Forge, Java, …) will be printed if present in the pack description
- The pack contents are present in `out/`
