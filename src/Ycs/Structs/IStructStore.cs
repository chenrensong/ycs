// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;

namespace Ycs.Structs
{
    public interface IStructStore
    {
        void AddStruct(AbstractStruct str);
        void CleanupPendingStructs();
        AbstractStruct Find(ID id);
        int FindIndexCleanStart(ITransaction transaction, List<AbstractStruct> structs, long clock);
        (AbstractStruct item, int diff) FollowRedone(ID id);
        AbstractStruct GetItemCleanEnd(ITransaction transaction, ID id);
        AbstractStruct GetItemCleanStart(ITransaction transaction, ID id);
        long GetState(long clientId);
        IDictionary<long, long> GetStateVector();
        void IntegrityCheck();
        void IterateStructs(ITransaction transaction, List<AbstractStruct> structs, long clockStart, long length, Predicate<AbstractStruct> fun);
        void ReadAndApplyDeleteSet(IDSDecoder decoder, ITransaction transaction);
        void ReplaceStruct(AbstractStruct oldStruct, AbstractStruct newStruct);
    }
}