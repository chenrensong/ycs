// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Diagnostics;
using Ycs.Contracts;

namespace Ycs.Structs
{
    public class GC : IItem
    {
        internal const byte StructGCRefNumber = 0;
        public StructID Id { get; set; }
        public int Length { get; set; }

        internal GC(StructID id, int length)
        {
            this.Id = id;
            this.Length = length;
            // Do nothing.
        }

        public bool Deleted => true;

        public IContentEx Content { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }
        public bool Countable { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }
        public bool Keep { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }

        public StructID LastId => throw new NotImplementedException();

        public IItem Left { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }
        public StructID? LeftOrigin { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }
        public bool Marker { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }

        public IItem Next => throw new NotImplementedException();

        public object Parent { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }
        public string ParentSub { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }

        public IItem Prev => throw new NotImplementedException();

        public StructID? Redone { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }
        public IItem Right { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }
        public StructID? RightOrigin { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }

        public bool MergeWith(IItem right)
        {
            Debug.Assert(right is GC);
            Length += right.Length;
            return true;
        }

        public void Delete(ITransaction transaction)
        {
            // Do nothing.
        }

        public  void Integrate(ITransaction transaction, int offset)
        {
            if (offset > 0)
            {
                Id = new StructID(Id.Client, Id.Clock + offset);
                Length -= offset;
            }

            transaction.Doc.Store.AddStruct(this);
        }

        public long? GetMissing(ITransaction transaction, IStructStore store)
        {
            return null;
        }

        public void Write(IUpdateEncoder encoder, int offset)
        {
            encoder.WriteInfo(StructGCRefNumber);
            encoder.WriteLength(Length - offset);
        }

        public void Gc(IStructStore store, bool parentGCd)
        {
            throw new NotImplementedException();
        }

        public bool IsVisible(ISnapshot snap)
        {
            throw new NotImplementedException();
        }

        public void KeepItemAndParents(bool value)
        {
            throw new NotImplementedException();
        }

        public void MarkDeleted()
        {
            throw new NotImplementedException();
        }

        public IItem SplitItem(ITransaction transaction, int diff)
        {
            throw new NotImplementedException();
        }
    }
}
