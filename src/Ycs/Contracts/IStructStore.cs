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
        IDictionary<long, List<AbstractStruct>> Clients { get; }
        void AddStruct(AbstractStruct str);
        void CleanupPendingStructs();
        AbstractStruct Find(StructID id);
        int FindIndexCleanStart(ITransaction transaction, List<AbstractStruct> structs, long clock);
        (AbstractStruct item, int diff) FollowRedone(StructID id);
        AbstractStruct GetItemCleanEnd(ITransaction transaction, StructID id);
        AbstractStruct GetItemCleanStart(ITransaction transaction, StructID id);
        long GetState(long clientId);
        IDictionary<long, long> GetStateVector();
        void IntegrityCheck();
        void IterateStructs(ITransaction transaction, List<AbstractStruct> structs, long clockStart, long length, Predicate<AbstractStruct> fun);
        void MergeReadStructsIntoPendingReads(IDictionary<long, List<AbstractStruct>> clientStructRefs);
        void ReadAndApplyDeleteSet(IDSDecoder decoder, ITransaction transaction);
        void ReplaceStruct(AbstractStruct oldStruct, AbstractStruct newStruct);
        void ResumeStructIntegration(ITransaction transaction);
        void TryResumePendingDeleteReaders(ITransaction transaction);
    }
}