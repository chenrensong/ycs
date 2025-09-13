// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System.Collections.Generic;
using Ycs.Structs;
using Ycs.Types;

namespace Ycs.Contracts
{
    public interface ITransaction
    {
        IDictionary<long, long> AfterState { get; set; }
        IDictionary<long, long> BeforeState { get; set; }
        IDictionary<AbstractType, ISet<string>> Changed { get; }
        IDictionary<AbstractType, IList<YEvent>> ChangedParentTypes { get; }
        DeleteSet DeleteSet { get; }
        YDoc Doc { get; }
        bool Local { get; }
        IDictionary<string, object> Meta { get; }
        object Origin { get; }
        ISet<YDoc> SubdocsAdded { get; }
        ISet<YDoc> SubdocsLoaded { get; }
        ISet<YDoc> SubdocsRemoved { get; }
        IList<AbstractStruct> MergeStructs { get; }

        void AddChangedTypeToTransaction(AbstractType type, string parentSub);
        StructID GetNextId();
        AbstractStruct RedoItem(Item item, ISet<Item> redoItems);
        bool WriteUpdateMessageFromTransaction(IUpdateEncoder encoder);
    }
}