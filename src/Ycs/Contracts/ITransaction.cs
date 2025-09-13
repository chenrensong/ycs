// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System.Collections.Generic;

namespace Ycs.Contracts
{
    public interface ITransaction
    {
        IDictionary<long, long> AfterState { get; set; }
        IDictionary<long, long> BeforeState { get; set; }
        IDictionary<IAbstractType, ISet<string>> Changed { get; }
        IDictionary<IAbstractType, IList<IYEvent>> ChangedParentTypes { get; }
        IDeleteSet DeleteSet { get; }
        IYDoc Doc { get; }
        bool Local { get; }
        IDictionary<string, object> Meta { get; }
        object Origin { get; }
        ISet<IYDoc> SubdocsAdded { get; }
        ISet<IYDoc> SubdocsLoaded { get; }
        ISet<IYDoc> SubdocsRemoved { get; }
        IList<IItem> MergeStructs { get; }
        void AddChangedTypeToTransaction(IAbstractType type, string parentSub);
        StructID GetNextId();
        IItem RedoItem(IItem item, ISet<IItem> redoItems);
        bool WriteUpdateMessageFromTransaction(IUpdateEncoder encoder);
    }
}