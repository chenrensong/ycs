// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;

namespace Ycs.Contracts
{
    public interface IStructStore
    {
        IDictionary<long, List<IStructItem>> Clients { get; }
        void AddStruct(IStructItem str);
        void CleanupPendingStructs();
        IStructItem Find(StructID id);
        int FindIndexCleanStart(ITransaction transaction, List<IStructItem> structs, long clock);
        (IStructItem item, int diff) FollowRedone(StructID id);
        IStructItem GetItemCleanEnd(ITransaction transaction, StructID id);
        IStructItem GetItemCleanStart(ITransaction transaction, StructID id);
        long GetState(long clientId);
        IDictionary<long, long> GetStateVector();
        void IntegrityCheck();
        void IterateStructs(ITransaction transaction, List<IStructItem> structs, long clockStart, long length, Predicate<IStructItem> fun);
        void MergeReadStructsIntoPendingReads(IDictionary<long, List<IStructItem>> clientStructRefs);
        void ReadAndApplyDeleteSet(IDSDecoder decoder, ITransaction transaction);
        void ReplaceStruct(IStructItem oldStruct, IStructItem newStruct);
        void ResumeStructIntegration(ITransaction transaction);
        void TryResumePendingDeleteReaders(ITransaction transaction);
    }
}