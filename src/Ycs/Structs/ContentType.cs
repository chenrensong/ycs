// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;
using Ycs.Types;
using Ycs.Contracts;

namespace Ycs.Structs
{
    public class ContentType : IContentEx
    {
        internal const int _ref = 7;

        internal ContentType(AbstractType type)
        {
            Type = type;
        }

        int IContentEx.Ref => _ref;

        public bool Countable => true;
        public int Length => 1;

        public AbstractType Type { get; }

        public IReadOnlyList<object> GetContent() => new object[] { Type };

        public IContent Copy() => new ContentType(Type.InternalCopy());

        public IContent Splice(int offset) => throw new NotImplementedException();

        public bool MergeWith(IContent right) => false;

        void IContentEx.Integrate(ITransaction transaction, Item item)
        {
            Type.Integrate(transaction.Doc, item);
        }

        void IContentEx.Delete(ITransaction transaction)
        {
            var item = Type.Start;

            while (item != null)
            {
                if (!item.Deleted)
                {
                    item.Delete(transaction);
                }
                else
                {
                    // This will be gc'd later and we want to merge it if possible.
                    // We try to merge all deleted items each transaction,
                    // but we have no knowledge about that this needs to merged
                    // since it is not in transaction. Hence we add it to transaction._mergeStructs.
                    transaction.MergeStructs.Add(item);
                }

                item = item.Right as Item;
            }

            foreach (var valueItem in Type.Map.Values)
            {
                if (!valueItem.Deleted)
                {
                    valueItem.Delete(transaction);
                }
                else
                {
                    // Same as above.
                    transaction.MergeStructs.Add(valueItem);
                }
            }

            transaction.Changed.Remove(Type);
        }

        void IContentEx.Gc(IStructStore store)
        {
            var item = Type.Start;
            while (item != null)
            {
                item.Gc(store, parentGCd: true);
                item = item.Right as Item;
            }

            Type.Start = null;

            foreach (var kvp in Type.Map)
            {
                var valueItem = kvp.Value;
                while (valueItem != null)
                {
                    valueItem.Gc(store, parentGCd: true);
                    valueItem = valueItem.Left as Item;
                }
            }

            Type.Map.Clear();
        }

        void IContentEx.Write(IUpdateEncoder encoder, int offset)
        {
            Type.Write(encoder);
        }

        internal static ContentType Read(IUpdateDecoder decoder)
        {
            var typeRef = decoder.ReadTypeRef();
            switch (typeRef)
            {
                case YArray.YArrayRefId:
                    var arr = YArray.Read(decoder);
                    return new ContentType(arr);
                case YMap.YMapRefId:
                    var map = YMap.Read(decoder);
                    return new ContentType(map);
                case YText.YTextRefId:
                    var text = YText.Read(decoder);
                    return new ContentType(text);
                default:
                    throw new NotImplementedException($"Type {typeRef} not implemented");
            }
        }
    }
}
