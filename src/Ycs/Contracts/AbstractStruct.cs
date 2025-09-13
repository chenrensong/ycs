// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Diagnostics;
using Ycs.Types;

namespace Ycs.Contracts
{
    public abstract class AbstractStruct
    {
        protected AbstractStruct(StructID id, int length)
        {
            Debug.Assert(length >= 0);

            Id = id;
            Length = length;
        }
        public StructID Id { get; set; }
        public int Length { get; set; }
        public abstract bool Deleted { get; }

        internal abstract bool MergeWith(AbstractStruct right);
        internal abstract void Delete(ITransaction transaction);
        internal abstract void Integrate(ITransaction transaction, int offset);
        internal abstract long? GetMissing(ITransaction transaction, IStructStore store);
        internal abstract void Write(IUpdateEncoder encoder, int offset);

     
    }
}
