using CUE4Parse.UE4.Assets.Exports.Texture;
using System.Buffers.Binary;
using System.IO.Compression;

// Narrow map-format adapter. Reassembles aligned compressed blocks before PNG decoding.
// CUE4Parse supplies archive/virtual-tile addressing; no arbitrary image-codec dependency.
internal static class BC1Export
{
    public static byte[] Read(UTexture2D texture)
    {
        if (texture.Format != EPixelFormat.PF_DXT1) throw new Exception("only DXT1 is supported");
        int width, height;
        byte[] blocks;
        var vt = texture.PlatformData.VTData;
        if (vt?.IsInitialized() == true)
        {
            width = checked((int)vt.Width); height = checked((int)vt.Height);
            var tile = checked((int)vt.TileSize); var border = checked((int)vt.TileBorderSize);
            if (vt.NumLayers != 1 || vt.LayerTypes[0] != EPixelFormat.PF_DXT1 ||
                tile < 4 || tile > 1024 || border < 0 || border > 32 || tile % 4 != 0 || border % 4 != 0 ||
                width % tile != 0 || height % tile != 0) throw new Exception("unsupported virtual BC1 layout");
            ValidateDimensions(width, height);
            blocks = new byte[(width / 4) * (height / 4) * 8];
            var physical = tile + 2 * border;
            var tileBytes = (physical / 4) * (physical / 4) * 8;
            var tileStride = (physical / 4) * 8;
            var imageStride = (width / 4) * 8;
            for (var y = 0; y < height / tile; y++)
            for (var x = 0; x < width / tile; x++)
            {
                var address = Morton((uint)x, (uint)y);
                if (!vt.IsValidAddress(0, address)) throw new Exception("virtual mip has missing tiles");
                if (!vt.IsLegacyData() && (vt.BaseOffsetPerMip?[0] == uint.MaxValue || vt.TileOffsetData?[0].GetTileOffset(address) == uint.MaxValue))
                    throw new Exception("virtual tile is not present");
                var (chunkIndex, start, length) = vt.GetTileData(0, address, 0);
                if (chunkIndex < 0 || chunkIndex >= vt.Chunks.Length) throw new Exception("invalid virtual tile chunk");
                var chunk = vt.Chunks[chunkIndex];
                var source = chunk.BulkData.Data ?? throw new Exception("missing virtual chunk data");
                var offset = checked((int)start); var bytes = checked((int)length);
                if (offset < 0 || bytes < 0 || offset > source.Length - bytes) throw new Exception("invalid tile bounds");
                byte[] compressed;
                switch (chunk.CodecType[0])
                {
                    case EVirtualTextureCodec.RawGPU:
                        // The pinned CUE4Parse returns layer offset, not byte length, for modern raw tiles.
                        if (offset > source.Length - tileBytes ||
                            (vt.IsLegacyData() && bytes < tileBytes)) throw new Exception("raw tile size mismatch");
                        compressed = source.AsSpan(offset, tileBytes).ToArray();
                        break;
                    case EVirtualTextureCodec.ZippedGPU_DEPRECATED:
                        compressed = new byte[tileBytes];
                        using (var input = new MemoryStream(source, offset, bytes, false))
                        using (var zip = new ZLibStream(input, CompressionMode.Decompress))
                        { zip.ReadExactly(compressed); if (zip.ReadByte() != -1) throw new Exception("oversized compressed tile"); }
                        break;
                    default: throw new Exception("virtual texture codec unsupported");
                }
                var borderBlocks = border / 4; var usefulBlocks = tile / 4;
                for (var row = 0; row < usefulBlocks; row++)
                {
                    var from = (row + borderBlocks) * tileStride + borderBlocks * 8;
                    var to = (y * usefulBlocks + row) * imageStride + x * usefulBlocks * 8;
                    compressed.AsSpan(from, usefulBlocks * 8).CopyTo(blocks.AsSpan(to, usefulBlocks * 8));
                }
            }
        }
        else
        {
            var mip = texture.GetFirstMip() ?? throw new Exception("texture mip unavailable");
            width = mip.SizeX; height = mip.SizeY;
            ValidateDimensions(width, height);
            if (mip.SizeZ != 1) throw new Exception("only 2D textures supported");
            blocks = mip.BulkData?.Data ?? throw new Exception("mip data unavailable");
            if (blocks.Length != ((width + 3) / 4) * ((height + 3) / 4) * 8) throw new Exception("mip size mismatch");
        }
        var result = new byte[16 + blocks.Length];
        System.Text.Encoding.ASCII.GetBytes("BWBC1\0\0\0").CopyTo(result, 0);
        BinaryPrimitives.WriteUInt32LittleEndian(result.AsSpan(8, 4), (uint)width);
        BinaryPrimitives.WriteUInt32LittleEndian(result.AsSpan(12, 4), (uint)height);
        blocks.CopyTo(result, 16);
        return result;
    }
    private static void ValidateDimensions(int width, int height)
    { if (width < 1 || height < 1 || width > 8192 || height > 8192) throw new Exception("unsupported bounded dimensions"); }
    private static uint Morton(uint x, uint y)
    {
        uint result = 0;
        for (int bit = 0; bit < 16; bit++) { result |= ((x >> bit) & 1) << (bit * 2); result |= ((y >> bit) & 1) << (bit * 2 + 1); }
        return result;
    }
}
