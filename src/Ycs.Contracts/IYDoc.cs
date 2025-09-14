using System;
using System.Collections.Generic;
using System.IO;

namespace Ycs.Contracts
{
    public interface IYDoc
    {
        bool AutoLoad { get; }
        int ClientId { get; set; }
        bool Gc { get; }
        Predicate<IStructItem> GcFilter { get; }
        string Guid { get; }
        IStructItem Item { get; set; }
        IDictionary<string, string> Meta { get; }
        YDocOptions Opts { get; }
        IDictionary<string, IAbstractType> Share { get; set; }
        bool ShouldLoad { get; set; }
        IStructStore Store { get; set; }
        ISet<IYDoc> Subdocs { get; }
        ITransaction Transaction { get; set; }
        IList<ITransaction> TransactionCleanups { get; }
     

        void ApplyUpdateV2(byte[] update, object transactionOrigin = null, bool local = false);
        void ApplyUpdateV2(Stream input, object transactionOrigin = null, bool local = false);
        YDocOptions CloneOptionsWithNewGuid();
        ISnapshot CreateSnapshot();
        void Destroy();
        byte[] EncodeStateAsUpdateV2(byte[] encodedTargetStateVector = null);
        byte[] EncodeStateVectorV2();
        string FindRootTypeKey(IAbstractType type);
        T Get<T>(string name) where T : IAbstractType, new();
        IYArray GetArray(string name = "");
        IYMap GetMap(string name = "");
        IEnumerable<string> GetSubdocGuids();
        IYText GetText(string name = "");
        void InvokeAfterAllTransactions(IList<ITransaction> transactions);
        void InvokeBeforeAllTransactions();
        void InvokeDestroyed();
        void InvokeOnAfterTransaction(ITransaction transaction);
        void InvokeOnAfterTransactionCleanup(ITransaction transaction);
        void InvokeOnBeforeObserverCalls(ITransaction transaction);
        void InvokeOnBeforeTransaction(ITransaction transaction);
        void InvokeSubdocsChanged(ISet<IYDoc> loaded, ISet<IYDoc> added, ISet<IYDoc> removed);
        void InvokeUpdateV2(ITransaction transaction);
        void Load();
        void Transact(Action<ITransaction> fun, object origin = null, bool local = true);
        void WriteStateAsUpdate(IUpdateEncoder encoder, IDictionary<long, long> targetStateVector);
        void WriteStateVector(IDSEncoder encoder);


        event EventHandler<ITransaction> BeforeObserverCalls;
        event EventHandler<ITransaction> BeforeTransaction;
        event EventHandler<ITransaction> AfterTransaction;
        event EventHandler<ITransaction> AfterTransactionCleanup;
        event EventHandler BeforeAllTransactions;
        event EventHandler<IList<ITransaction>> AfterAllTransactions;
        event EventHandler<(byte[] data, object origin, ITransaction transaction)> UpdateV2;
        event EventHandler Destroyed;
        event EventHandler<(ISet<IYDoc> Loaded, ISet<IYDoc> Added, ISet<IYDoc> Removed)> SubdocsChanged;

    }
}