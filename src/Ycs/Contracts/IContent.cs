// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System.Collections.Generic;
using Ycs.Structs;

namespace Ycs.Contracts
{
    public interface IContent
    {
        bool Countable { get; }
        int Length { get; }
        IReadOnlyList<object> GetContent();
        IContent Copy();
        IContent Splice(int offset);
        bool MergeWith(IContent right);
    }

    internal interface IContentEx : IContent
    {
        int Ref { get; }

        void Integrate(ITransaction transaction, Item item);
        void Delete(ITransaction transaction);
        void Gc(IStructStore store);
        void Write(IUpdateEncoder encoder, int offset);
    }
}
