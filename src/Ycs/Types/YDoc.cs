// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using Ycs.Structs;
using Ycs.Contracts;
using Ycs.Utils;

namespace Ycs.Types
{

    /// <summary>
    /// Yjs instance handles the state of shared data.
    /// </summary>
    public class YDoc : IYDoc
    {
        public YDocOptions Opts { get; }
        public string Guid => Opts.Guid;
        public bool Gc => Opts.Gc;
        public Predicate<IItem> GcFilter => Opts.GcFilter;
        public bool AutoLoad => Opts.AutoLoad;
        public IDictionary<string, string> Meta => Opts.Meta;

        public bool ShouldLoad { get; set; }
        public IList<ITransaction> TransactionCleanups { get; }
        public ITransaction Transaction { get; set; }
        public ISet<IYDoc> Subdocs { get; }
        // If this document is a subdocument - a document integrated into another document - them _item is defined.
        public IItem Item { get; set; }
        public IDictionary<string, IAbstractType> Share { get; set; }

        public static int GenerateNewClientId()
        {
            return new Random().Next(0, int.MaxValue);
        }

        /// <param name="gc">Disable garbage collection.</param>
        /// <param name="gcFilter">WIll be called before an Item is garbage collected. Return false to keep the item.</param>
        public YDoc(YDocOptions opts = null)
        {
            Opts = opts ?? new YDocOptions();
            TransactionCleanups = new List<ITransaction>();

            ClientId = GenerateNewClientId();
            Share = new Dictionary<string, IAbstractType>();
            Store = new StructStore();
            Subdocs = new HashSet<IYDoc>();
            ShouldLoad = Opts.AutoLoad;
        }

        /// <summary>
        /// Notify the parent document that you request to load data into this subdocument (if it is a subdocument).
        /// 'load()' might be used in the future to request any provider to load the most current data.
        /// It is safe to call 'Load()' multiple times.
        /// </summary>
        public void Load()
        {
            var item = Item;
            if (item != null && !ShouldLoad)
            {
                Debug.Assert(item.Parent is AbstractType);
                (item.Parent as AbstractType).Doc.Transact(tr =>
                {
                    tr.SubdocsLoaded.Add(this);
                }, origin: null, local: true);
            }
            ShouldLoad = true;
        }

        public ISnapshot CreateSnapshot() => new Snapshot(new DeleteSet(Store), Store.GetStateVector());

        public IEnumerable<string> GetSubdocGuids()
        {
            return new HashSet<string>(Subdocs.Select(sd => sd.Guid));
        }

        public void Destroy()
        {
            foreach (var sd in Subdocs)
            {
                sd.Destroy();
            }

            var item = Item;
            if (item != null)
            {
                Item = null;
                var content = item.Content as ContentDoc;

                if (item.Deleted)
                {
                    if (content != null)
                    {
                        content.Doc = null;
                    }
                }
                else
                {
                    Debug.Assert(content != null);
                    var newOpts = content.Opts;
                    newOpts.Guid = Guid;

                    content.Doc = new YDoc(newOpts);
                    content.Doc.Item = item;
                }

                (item.Parent as AbstractType).Doc.Transact(tr =>
                {
                    if (!item.Deleted)
                    {
                        Debug.Assert(content != null);
                        tr.SubdocsAdded.Add(content.Doc);
                    }

                    tr.SubdocsRemoved.Add(this);
                }, origin: null, local: true);
            }

            InvokeDestroyed();
        }

        public event EventHandler<ITransaction> BeforeObserverCalls;
        public event EventHandler<ITransaction> BeforeTransaction;
        public event EventHandler<ITransaction> AfterTransaction;
        public event EventHandler<ITransaction> AfterTransactionCleanup;
        public event EventHandler BeforeAllTransactions;
        public event EventHandler<IList<ITransaction>> AfterAllTransactions;
        public event EventHandler<(byte[] data, object origin, ITransaction transaction)> UpdateV2;
        public event EventHandler Destroyed;
        public event EventHandler<(ISet<IYDoc> Loaded, ISet<IYDoc> Added, ISet<IYDoc> Removed)> SubdocsChanged;

        public int ClientId { get; set; }
        public IStructStore Store { get; set; }

        /// <summary>
        /// Changes that happen inside of a transaction are bundled.
        /// This means that the observer fires _after_ the transaction is finished and that
        /// all changes that happened inside of the transaction are sent as one message to the
        /// other peers.
        /// </summary>
        /// <param name="fun">The function that should be executed as a transaction.</param>
        /// <param name="origin">ITransaction owner. Will be stored in 'transaction.origin'.</param>
        /// <param name="local"></param>
        public void Transact(Action<ITransaction> fun, object origin = null, bool local = true)
        {
            bool initialCall = false;
            if (Transaction == null)
            {
                initialCall = true;
                Transaction = new Transaction(this, origin, local);
                TransactionCleanups.Add(Transaction);
                if (TransactionCleanups.Count == 1)
                {
                    InvokeBeforeAllTransactions();
                }

                InvokeOnBeforeTransaction(Transaction);
            }

            try
            {
                fun(Transaction);
            }
            finally
            {
                if (initialCall && TransactionCleanups[0] == Transaction)
                {
                    // The first transaction ended, now process observer calls.
                    // Observer call may create new transacations for which we need to call the observers and do cleanup.
                    // We don't want to nest these calls, so we execute these calls one after another.
                    // Also we need to ensure that all cleanups are called, even if the observers throw errors.
                    Utils.Transaction.CleanupTransactions(TransactionCleanups, 0);
                }
            }
        }

        public IYArray GetArray(string name = "")
        {
            return Get<YArray>(name);
        }

        public IYMap GetMap(string name = "")
        {
            return Get<YMap>(name);
        }

        public IYText GetText(string name = "")
        {
            return Get<YText>(name);
        }

        public T Get<T>(string name)
            where T : IAbstractType, new()
        {
            if (!Share.TryGetValue(name, out var type))
            {
                type = new T();
                type.Integrate(this, null);
                Share[name] = type;
            }

            // Remote type is realized when this method is called.
            if (typeof(T) != typeof(AbstractType) && !typeof(T).IsAssignableFrom(type.GetType()))
            {
                if (type.GetType() == typeof(AbstractType))
                {
                    var t = new T();
                    t.Map = type.Map;

                    foreach (var kvp in type.Map)
                    {
                        var n = kvp.Value;
                        for (; n != null; n = n.Left as IItem)
                        {
                            n.Parent = t;
                        }
                    }

                    t.Start = type.Start;
                    for (var n = t.Start; n != null; n = n.Right as IItem)
                    {
                        n.Parent = t;
                    }

                    t.Length = type.Length;

                    Share[name] = t;
                    t.Integrate(this, null);
                    return t;
                }
                else
                {
                    throw new Exception($"Type with the name {name} has already been defined with a different constructor");
                }
            }

            return (T)type;
        }

        /// <summary>
        /// Read and apply a document update.
        /// <br/>
        /// This function has the same effect as 'applyUpdate' but accepts a decoder.
        /// </summary>
        public void ApplyUpdateV2(Stream input, object transactionOrigin = null, bool local = false)
        {
            Transact(tr =>
            {
                using (var structDecoder = new UpdateDecoderV2(input))
                {
                    EncodingUtils.ReadStructs(structDecoder, tr, Store);
                    Store.ReadAndApplyDeleteSet(structDecoder, tr);
                }
            }, transactionOrigin, local);
        }

        public void ApplyUpdateV2(byte[] update, object transactionOrigin = null, bool local = false)
        {
            using (var input = new MemoryStream(update, writable: false))
            {
                ApplyUpdateV2(input, transactionOrigin, local);
            }
        }

        /// <summary>
        /// Write all the document as a single update message that can be applied on the remote document.
        /// If you specify the state of the remote client, it will only write the operations that are missing.
        /// <br/>
        /// Use 'WriteStateAsUpdate' instead if you are working with Encoder.
        /// </summary>
        public byte[] EncodeStateAsUpdateV2(byte[] encodedTargetStateVector = null)
        {
            using (var encoder = new UpdateEncoderV2())
            {
                var targetStateVector = encodedTargetStateVector == null
                    ? new Dictionary<long, long>()
                    : EncodingUtils.DecodeStateVector(new MemoryStream(encodedTargetStateVector, writable: false));
                WriteStateAsUpdate(encoder, targetStateVector);
                return encoder.ToArray();
            }
        }

        public byte[] EncodeStateVectorV2()
        {
            using (var encoder = new DSEncoderV2())
            {
                WriteStateVector(encoder);
                return encoder.ToArray();
            }
        }

        /// <summary>
        /// Write all the document as a single update message. If you specify the satte of the remote client, it will only
        /// write the operations that are missing.
        /// </summary>
        public void WriteStateAsUpdate(IUpdateEncoder encoder, IDictionary<long, long> targetStateVector)
        {
            EncodingUtils.WriteClientsStructs(encoder, Store, targetStateVector);
            new DeleteSet(Store).Write(encoder);
        }

        public void WriteStateVector(IDSEncoder encoder)
        {
            EncodingUtils.WriteStateVector(encoder, Store.GetStateVector());
        }

        public void InvokeSubdocsChanged(ISet<IYDoc> loaded, ISet<IYDoc> added, ISet<IYDoc> removed)
        {
            SubdocsChanged?.Invoke(this, (loaded, added, removed));
        }

        public void InvokeOnBeforeObserverCalls(ITransaction transaction)
        {
            BeforeObserverCalls?.Invoke(this, transaction);
        }

        public void InvokeAfterAllTransactions(IList<ITransaction> transactions)
        {
            AfterAllTransactions?.Invoke(this, transactions);
        }

        public void InvokeOnBeforeTransaction(ITransaction transaction)
        {
            BeforeTransaction?.Invoke(this, transaction);
        }

        public void InvokeOnAfterTransaction(ITransaction transaction)
        {
            AfterTransaction?.Invoke(this, transaction);
        }

        public void InvokeOnAfterTransactionCleanup(ITransaction transaction)
        {
            AfterTransactionCleanup?.Invoke(this, transaction);
        }

        public void InvokeBeforeAllTransactions()
        {
            BeforeAllTransactions?.Invoke(this, null);
        }

        public void InvokeDestroyed()
        {
            Destroyed?.Invoke(this, null);
        }

        public void InvokeUpdateV2(ITransaction transaction)
        {
            var handler = UpdateV2;
            if (handler != null)
            {
                using (var encoder = new UpdateEncoderV2())
                {
                    var hasContent = transaction.WriteUpdateMessageFromTransaction(encoder);
                    if (hasContent)
                    {
                        handler.Invoke(this, (encoder.ToArray(), transaction.Origin, transaction));
                    }
                }
            }
        }

        public YDocOptions CloneOptionsWithNewGuid()
        {
            var newOpts = Opts.Clone();
            newOpts.Guid = System.Guid.NewGuid().ToString("D");
            return newOpts;
        }

        public string FindRootTypeKey(IAbstractType type)
        {
            foreach (var kvp in Share)
            {
                if (type?.Equals(kvp.Value) ?? false)
                {
                    return kvp.Key;
                }
            }

            throw new Exception();
        }
    }
}
