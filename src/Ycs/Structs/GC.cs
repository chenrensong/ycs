// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System.Diagnostics;

namespace Ycs.Structs
{
    public class GC : AbstractStruct
    {
        internal const byte StructGCRefNumber = 0;

        internal GC(ID id, int length)
            : base(id, length)
        {
            // Do nothing.
        }

        public override bool Deleted => true;

        internal override bool MergeWith(AbstractStruct right)
        {
            Debug.Assert(right is GC);
            Length += right.Length;
            return true;
        }

        internal override void Delete(ITransaction transaction)
        {
            // Do nothing.
        }

        internal override void Integrate(ITransaction transaction, int offset)
        {
            if (offset > 0)
            {
                Id = new ID(Id.Client, Id.Clock + offset);
                Length -= offset;
            }

            transaction.Doc.Store.AddStruct(this);
        }

        internal override long? GetMissing(ITransaction transaction, IStructStore store)
        {
            return null;
        }

        internal override void Write(IUpdateEncoder encoder, int offset)
        {
            encoder.WriteInfo(StructGCRefNumber);
            encoder.WriteLength(Length - offset);
        }
    }
}
