// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;

namespace Ycs.Contracts
{
    public interface IDeleteSet
    {
        IDictionary<long, List<DeleteItem>> Clients { get; }
        void Add(long client, long clock, long length);
        int? FindIndexSS(IList<DeleteItem> dis, long clock);
        bool IsDeleted(StructID id);
        void IterateDeletedStructs(ITransaction transaction, Predicate<IStructItem> fun);
        void SortAndMergeDeleteSet();
        void TryGc(IStructStore store, Predicate<IStructItem> gcFilter);
        void TryGcDeleteSet(IStructStore store, Predicate<IStructItem> gcFilter);
        void TryMergeDeleteSet(IStructStore store);
        void Write(IDSEncoder encoder);
    }
}