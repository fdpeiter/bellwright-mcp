// Steward's local single-package extraction helper. Uses Apache-2.0 CUE4Parse.
// Archives/mappings are owner-supplied; no key submission, game attachment or asset download.
using CUE4Parse.FileProvider;
using CUE4Parse.UE4.Versions;
using CUE4Parse.UE4.Assets.Exports.Texture;
using CUE4Parse.MappingsProvider.Usmap;
using Newtonsoft.Json;
using System.Security.Cryptography;

if (args.Length != 5 || args[4] is not ("texture" or "json"))
{
    Console.Error.WriteLine("usage: AssetExtract PAK_DIRECTORY MAPPINGS.usmap PACKAGE_SUFFIX OUTPUT_FILE texture|json");
    return 2;
}
try
{
    var archiveDir = Path.GetFullPath(args[0]);
    var output = Path.GetFullPath(args[3]);
    if (!Directory.Exists(archiveDir) || !File.Exists(args[1])) throw new Exception("owner archives/mappings missing");
    if (output.StartsWith(archiveDir + Path.DirectorySeparatorChar, StringComparison.OrdinalIgnoreCase))
        throw new Exception("output must be outside the game archive directory");
    if (File.Exists(output) || File.Exists(output + ".provenance.json")) throw new Exception("output exists; choose a new private output path");
    var version = Enum.Parse<EGame>(Environment.GetEnvironmentVariable("GAME") ?? "GAME_UE5_7");
    var provider = new DefaultFileProvider(archiveDir, SearchOption.TopDirectoryOnly, new VersionContainer(version));
    provider.MappingsContainer = new FileUsmapTypeMappingsProvider(args[1]);
    provider.Initialize();
    provider.Mount();
    provider.PostMount();
    var suffix = args[2].Replace('\\', '/').TrimStart('/');
    var candidates = provider.Files.Keys.Where(k => k.Equals(suffix, StringComparison.OrdinalIgnoreCase) ||
        k.EndsWith('/' + suffix, StringComparison.OrdinalIgnoreCase)).ToArray();
    if (candidates.Length != 1) throw new Exception("package suffix must identify exactly one mounted package; include extension and more path segments");
    var exports = provider.LoadPackage(candidates[0]).GetExports().ToArray();
    byte[] content;
    if (args[4] == "texture")
    {
        var textures = exports.OfType<UTexture2D>().ToArray();
        if (textures.Length != 1) throw new Exception("package must have exactly one Texture2D");
        content = BC1Export.Read(textures[0]);
    }
    else content = System.Text.Encoding.UTF8.GetBytes(JsonConvert.SerializeObject(exports, Formatting.None));
    Directory.CreateDirectory(Path.GetDirectoryName(output)!);
    using (var file = new FileStream(output, FileMode.CreateNew, FileAccess.Write)) file.Write(content);
    var provenance = new {
        source_kind = "packaged_game_asset", package = candidates[0], engine = version.ToString(),
        extractor = "Steward AssetExtract", cue4parse = "1.2.2.202610", extracted_at = DateTimeOffset.UtcNow,
        mappings_sha256 = Convert.ToHexString(SHA256.HashData(File.ReadAllBytes(args[1]))).ToLowerInvariant(),
        output_sha256 = Convert.ToHexString(SHA256.HashData(content)).ToLowerInvariant(),
        unknowns = new[] { "Game build and package hashes must be recorded by the owner.", "Decoded exports do not establish current runtime resource availability or world coverage." }
    };
    using (var file = new FileStream(output + ".provenance.json", FileMode.CreateNew, FileAccess.Write))
    using (var writer = new StreamWriter(file)) writer.Write(JsonConvert.SerializeObject(provenance, Formatting.Indented));
    Console.WriteLine("Extracted one owner-supplied package; output and provenance remain private.");
    return 0;
}
catch (Exception error)
{
    // Optional diagnostics are private owner output, never printed or attached to issues.
    var diagnostic = Environment.GetEnvironmentVariable("STEWARD_EXTRACT_DIAGNOSTICS");
    if (!string.IsNullOrWhiteSpace(diagnostic))
    {
        try
        {
        var path = Path.GetFullPath(diagnostic);
        var archives = Path.GetFullPath(args[0]);
        if (!path.StartsWith(archives + Path.DirectorySeparatorChar, StringComparison.OrdinalIgnoreCase))
        {
            Directory.CreateDirectory(Path.GetDirectoryName(path)!);
            File.WriteAllText(path, error.ToString());
        }
        }
        catch { /* Diagnostic failure must not expose exception details on stderr. */ }
    }
    Console.Error.WriteLine("Extraction failed. Verify authorized inputs, unique package suffix, engine/mappings compatibility and a new private output path.");
    return 1;
}
