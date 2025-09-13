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
        IDictionary<long, List<IItem>> Clients { get; }
        void AddStruct(IItem str);
        void CleanupPendingStructs();
        IItem Find(StructID id);
        int FindIndexCleanStart(ITransaction transaction, List<IItem> structs, long clock);
        (IItem item, int diff) FollowRedone(StructID id);
        IItem GetItemCleanEnd(ITransaction transaction, StructID id);
        IItem GetItemCleanStart(ITransaction transaction, StructID id);
        long GetState(long clientId);
        IDictionary<long, long> GetStateVector();
        void IntegrityCheck();
        void IterateStructs(ITransaction transaction, List<IItem> structs, long clockStart, long length, Predicate<IItem> fun);
        void MergeReadStructsIntoPendingReads(IDictionary<long, List<IItem>> clientStructRefs);
        void ReadAndApplyDeleteSet(IDSDecoder decoder, ITransaction transaction);
        void ReplaceStruct(IItem oldStruct, IItem newStruct);
        void ResumeStructIntegration(ITransaction transaction);
        void TryResumePendingDeleteReaders(ITransaction transaction);
    }
}