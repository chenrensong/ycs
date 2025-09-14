
namespace Ycs.Contracts
{
    public interface IStructItem
    {
        bool MergeWith(IStructItem right);
        void Delete(ITransaction transaction);
        void Integrate(ITransaction transaction, int offset);
        long? GetMissing(ITransaction transaction, IStructStore store);
        void Write(IUpdateEncoder encoder, int offset);
        StructID Id { get; set; }
        IContentEx Content { get; set; }
        bool Countable { get; set; }
        bool Deleted { get; }
        bool Keep { get; set; }
        StructID LastId { get; }
        IStructItem Left { get; set; }
        StructID? LeftOrigin { get; set; }
        bool Marker { get; set; }
        IStructItem Next { get; }
        object Parent { get; set; }
        string ParentSub { get; set; }
        IStructItem Prev { get; }
        StructID? Redone { get; set; }
        IStructItem Right { get; set; }
        StructID? RightOrigin { get; set; }
        int Length { get; set; }

        void Gc(IStructStore store, bool parentGCd);
        bool IsVisible(ISnapshot snap);
        void KeepItemAndParents(bool value);
        void MarkDeleted();
        IStructItem SplitItem(ITransaction transaction, int diff);
    }
}