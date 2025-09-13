﻿// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using Ycs.Content;
using Ycs.Contracts;
using Ycs.Types;

namespace Ycs.Core
{
    /// <summary>
    /// A transaction is created for every change on the Yjs model. It is possible
    /// to bundle changes on the Yjs model in a single transaction to minimize
    /// the number of messages sent and the number of observer calls.
    /// If possible the user of this library should bundle as many changes as possible.
    /// </summary>
    public class Transaction : ITransaction
    {
        // TODO: [alekseyk] To private?
        public readonly IList<IItem> _mergeStructs;

        public IList<IItem> MergeStructs => _mergeStructs;

        public Transaction(IYDoc doc, object origin, bool local)
        {
            Doc = doc;
            DeleteSet = new DeleteSet();
            BeforeState = Doc.Store.GetStateVector();
            AfterState = new Dictionary<long, long>();
            Changed = new Dictionary<IAbstractType, ISet<string>>();
            ChangedParentTypes = new Dictionary<IAbstractType, IList<IYEvent>>();
            _mergeStructs = new List<IItem>();
            Origin = origin;
            Meta = new Dictionary<string, object>();
            Local = local;
            SubdocsAdded = new HashSet<IYDoc>();
            SubdocsRemoved = new HashSet<IYDoc>();
            SubdocsLoaded = new HashSet<IYDoc>();
        }

        /// <summary>
        /// The Yjs instance.
        /// </summary>
        public IYDoc Doc { get; }

        public object Origin { get; }

        /// <summary>
        /// Holds the state before the transaction started.
        /// </summary>
        public IDictionary<long, long> BeforeState { get; set; }

        /// <summary>
        /// Holds the state after the transaction.
        /// </summary>
        public IDictionary<long, long> AfterState { get; set; }

        /// <summary>
        /// All types that were directly modified (property added or child
        /// inserted/deleted). New types are not included in this Set.
        /// Maps from type to parentSubs ('item.parentSub = null' for YArray).
        /// </summary>
        public IDictionary<IAbstractType, ISet<string>> Changed { get; }

        /// <summary>
        /// Stores the events for the types that observe also child elements.
        /// It is mainly used by 'observeDeep'.
        /// </summary>
        public IDictionary<IAbstractType, IList<IYEvent>> ChangedParentTypes { get; }

        /// <summary>
        /// Stores meta information on the transaction.
        /// </summary>
        public IDictionary<string, object> Meta { get; }

        /// <summary>
        /// Whether this change originates from this doc.
        /// </summary>
        public bool Local { get; }

        public ISet<IYDoc> SubdocsAdded { get; }

        public ISet<IYDoc> SubdocsRemoved { get; }

        public ISet<IYDoc> SubdocsLoaded { get; }

        /// <summary>
        /// Describes the set of deleted items by Ids.
        /// </summary>
        public IDeleteSet DeleteSet { get; }

        public StructID GetNextId()
        {
            return new StructID(Doc.ClientId, Doc.Store.GetState(Doc.ClientId));
        }

        /// <summary>
        /// If 'type.parent' was added in current transaction, 'type' technically did not change,
        /// it was just added and we should not fire events for 'type'.
        /// </summary>
        public void AddChangedTypeToTransaction(IAbstractType type, string parentSub)
        {
            var item = type.Item;
            if (item == null || BeforeState.TryGetValue(item.Id.Client, out var clock) && item.Id.Clock < clock && !item.Deleted)
            {
                if (!Changed.TryGetValue(type, out var set))
                {
                    set = new HashSet<string>();
                    Changed[type] = set;
                }

                set.Add(parentSub);
            }
        }

        public static void CleanupTransactions(IList<ITransaction> transactionCleanups, int i)
        {
            if (i < transactionCleanups.Count)
            {
                var transaction = transactionCleanups[i];
                var doc = transaction.Doc;
                var store = doc.Store;
                var ds = transaction.DeleteSet;
                var mergeStructs = transaction.MergeStructs;
                var actions = new List<Action>();

                try
                {
                    ds.SortAndMergeDeleteSet();
                    transaction.AfterState = store.GetStateVector();
                    doc.Transaction = null;

                    actions.Add(() =>
                    {
                        doc.InvokeOnBeforeObserverCalls(transaction);
                    });

                    actions.Add(() =>
                    {
                        foreach (var kvp in transaction.Changed)
                        {
                            var itemType = kvp.Key;
                            var subs = kvp.Value;

                            if (itemType.Item == null || !itemType.Item.Deleted)
                            {
                                itemType.CallObserver(transaction, subs);
                            }
                        }
                    });

                    actions.Add(() =>
                    {
                        // Deep observe events.
                        foreach (var kvp in transaction.ChangedParentTypes)
                        {
                            var type = kvp.Key;
                            var events = kvp.Value;

                            // We need to think about the possibility that the user transforms the YDoc in the event.
                            if (type.Item == null || !type.Item.Deleted)
                            {
                                foreach (var evt in events)
                                {
                                    if (evt.Target.Item == null || !evt.Target.Item.Deleted)
                                    {
                                        evt.CurrentTarget = type;
                                    }
                                }

                                // Sort events by path length so that top-level events are fired first.
                                var sortedEvents = events.ToList();
                                sortedEvents.Sort((a, b) => a.Path.Count - b.Path.Count);
                                Debug.Assert(sortedEvents.Count > 0);

                                actions.Add(() =>
                                {
                                    type.CallDeepEventHandlerListeners(sortedEvents, transaction);
                                });
                            }
                        }
                    });

                    actions.Add(() =>
                    {
                        doc.InvokeOnAfterTransaction(transaction);
                    });

                    CallAll(actions);
                }
                finally
                {
                    // Replace deleted items with ItemDeleted / GC.
                    // This is where content is actually removed from the Yjs Doc.
                    if (doc.Gc)
                    {
                        ds.TryGcDeleteSet(store, doc.GcFilter);
                    }

                    ds.TryMergeDeleteSet(store);

                    // On all affected store.clients props, try to merge.
                    foreach (var kvp in transaction.AfterState)
                    {
                        var client = kvp.Key;
                        var clock = kvp.Value;

                        if (!transaction.BeforeState.TryGetValue(client, out var beforeClock))
                        {
                            beforeClock = 0;
                        }

                        if (beforeClock != clock)
                        {
                            var structs = store.Clients[client];
                            var firstChangePos = Math.Max(StructStore.FindIndexSS(structs, beforeClock), 1);
                            for (int j = structs.Count - 1; j >= firstChangePos; j--)
                            {
                                Core.DeleteSet.TryToMergeWithLeft(structs, j);
                            }
                        }
                    }

                    // Try to merge mergeStructs.
                    // TODO: It makes more sense to transform mergeStructs to a DS, sort it, and merge from right to left
                    //       but at the moment DS does not handle duplicates.
                    for (int j = 0; j < mergeStructs.Count; j++)
                    {
                        var client = mergeStructs[j].Id.Client;
                        var clock = mergeStructs[j].Id.Clock;
                        var structs = store.Clients[client];
                        var replacedStructPos = StructStore.FindIndexSS(structs, clock);

                        if (replacedStructPos + 1 < structs.Count)
                        {
                            Core.DeleteSet.TryToMergeWithLeft(structs, replacedStructPos + 1);
                        }

                        if (replacedStructPos > 0)
                        {
                            Core.DeleteSet.TryToMergeWithLeft(structs, replacedStructPos);
                        }
                    }

                    if (!transaction.Local)
                    {
                        if (!transaction.AfterState.TryGetValue(doc.ClientId, out var afterClock))
                        {
                            afterClock = -1;
                        }

                        if (!transaction.BeforeState.TryGetValue(doc.ClientId, out var beforeClock))
                        {
                            beforeClock = -1;
                        }

                        if (afterClock != beforeClock)
                        {
                            doc.ClientId = YDoc.GenerateNewClientId();
                            // Debug.WriteLine($"{nameof(Transaction)}: Changed the client-id because another client seems to be using it.");
                        }
                    }

                    // @todo: Merge all the transactions into one and provide send the data as a single update message.
                    doc.InvokeOnAfterTransactionCleanup(transaction);

                    doc.InvokeUpdateV2(transaction);

                    foreach (var subDoc in transaction.SubdocsAdded)
                    {
                        doc.Subdocs.Add(subDoc);
                    }

                    foreach (var subDoc in transaction.SubdocsRemoved)
                    {
                        doc.Subdocs.Remove(subDoc);
                    }

                    doc.InvokeSubdocsChanged(transaction.SubdocsLoaded, transaction.SubdocsAdded, transaction.SubdocsRemoved);

                    foreach (var subDoc in transaction.SubdocsRemoved)
                    {
                        subDoc.Destroy();
                    }

                    if (transactionCleanups.Count <= i + 1)
                    {
                        doc.TransactionCleanups.Clear();
                        doc.InvokeAfterAllTransactions(transactionCleanups);
                    }
                    else
                    {
                        CleanupTransactions(transactionCleanups, i + 1);
                    }
                }
            }
        }

        /// <summary>
        /// Redoes the effect of this operation.
        /// </summary>
        public IItem RedoItem(IItem item, ISet<IItem> redoItems)
        {
            var doc = Doc;
            var store = doc.Store;
            var ownClientId = doc.ClientId;
            var redone = item.Redone;

            if (redone != null)
            {
                return store.GetItemCleanStart(this, redone.Value);
            }

            var parentItem = (item.Parent as AbstractType)?.Item;
            IItem left;
            IItem right;

            if (item.ParentSub == null)
            {
                // Is an array item. Insert at the old position.
                left = item.Left;
                right = item;
            }
            else
            {
                // Is a map item. Insert at current value.
                left = item;
                while ((left as IItem)?.Right != null)
                {
                    left = (left as IItem).Right;
                    if (left.Id.Client != ownClientId)
                    {
                        // It is not possible to redo this item because it conflicts with a change from another client.
                        return null;
                    }
                }

                if ((left as IItem)?.Right != null)
                {
                    left = (item.Parent as AbstractType)?.Map[item.ParentSub];
                }

                right = null;
            }

            // Make sure that parent is redone.
            if (parentItem != null && parentItem.Deleted && parentItem.Redone == null)
            {
                // Try to undo parent if it will be undone anyway.
                if (!redoItems.Contains(parentItem) || RedoItem(parentItem, redoItems) == null)
                {
                    return null;
                }
            }

            if (parentItem != null && parentItem.Redone != null)
            {
                while (parentItem.Redone != null)
                {
                    parentItem = store.GetItemCleanStart(this, parentItem.Redone.Value);
                }

                // Find next cloned_redo items.
                while (left != null)
                {
                    var leftTrace = left;
                    while (leftTrace != null && ((leftTrace as IItem)?.Parent as AbstractType)?.Item != parentItem)
                    {
                        leftTrace = (leftTrace as IItem).Redone == null ? null : store.GetItemCleanStart(this, (leftTrace as IItem).Redone.Value);
                    }

                    if (leftTrace != null && ((leftTrace as IItem)?.Parent as AbstractType)?.Item == parentItem)
                    {
                        left = leftTrace;
                        break;
                    }

                    left = (left as IItem)?.Left;
                }

                while (right != null)
                {
                    var rightTrace = right;
                    while (rightTrace != null && ((rightTrace as IItem)?.Parent as AbstractType)?.Item != parentItem)
                    {
                        rightTrace = (rightTrace as IItem).Redone == null ? null : store.GetItemCleanStart(this, (rightTrace as IItem).Redone.Value);
                    }

                    if (rightTrace != null && ((rightTrace as IItem)?.Parent as AbstractType)?.Item == parentItem)
                    {
                        right = rightTrace;
                        break;
                    }

                    right = (right as IItem)?.Right;
                }
            }

            var nextClock = store.GetState(ownClientId);
            var nextId = new StructID(ownClientId, nextClock);

            var redoneItem = new StructItem(
                nextId,
                left,
                (left as IItem)?.LastId,
                right,
                right?.Id,
                parentItem == null ? item.Parent : (parentItem.Content as ContentType)?.Type,
                item.ParentSub,
                item.Content.Copy());

            item.Redone = nextId;

            redoneItem.KeepItemAndParents(true);
            redoneItem.Integrate(this, 0);

            return redoneItem;
        }

        /// <returns>Whether the data was written.</returns>
        public bool WriteUpdateMessageFromTransaction(IUpdateEncoder encoder)
        {
            if (DeleteSet.Clients.Count == 0 && !AfterState.Any(kvp => !BeforeState.TryGetValue(kvp.Key, out var clockB) || kvp.Value != clockB))
            {
                return false;
            }

            DeleteSet.SortAndMergeDeleteSet();
            EncodingUtils.WriteClientsStructs(encoder, Doc.Store, BeforeState);
            DeleteSet.Write(encoder);

            return true;
        }

        private static void CallAll(IList<Action> funcs, int index = 0)
        {
            try
            {
                for (; index < funcs.Count; index++)
                {
                    funcs[index]();
                }
            }
            finally
            {
                if (index < funcs.Count)
                {
                    CallAll(funcs, index + 1);
                }
            }
        }
    }
}
