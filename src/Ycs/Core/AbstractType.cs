// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;
using System.Linq;
using Ycs.Contracts;
using Ycs.Types;
using Ycs.Content;

namespace Ycs.Core
{
    public class YEventArgs
    {
        public YEventArgs(IYEvent evt, ITransaction transaction)
        {
            Event = evt;
            Transaction = transaction;
        }

        public IYEvent Event { get; }
        public ITransaction Transaction { get; }
    }

    public class YDeepEventArgs
    {
        public YDeepEventArgs(IList<IYEvent> events, ITransaction transaction)
        {
            Events = events;
            Transaction = transaction;
        }

        public IList<IYEvent> Events { get; }
        public ITransaction Transaction { get; }
    }

    public class AbstractType : IAbstractType
    {
        public IItem Item { get; set; }
        public IItem Start { get; set; }
        public IDictionary<string, IItem> Map { get; set; } = new Dictionary<string, IItem>();

        public event EventHandler<YEventArgs> EventHandler;
        public event EventHandler<YDeepEventArgs> DeepEventHandler;

        public IYDoc Doc { get; protected set; }
        public IAbstractType Parent => Item != null ? Item.Parent as AbstractType : null;

        public virtual int Length { get; set; }

        public virtual void Integrate(IYDoc doc, IItem item)
        {
            Doc = doc;
            Item = item;
        }
        public virtual IAbstractType InternalCopy() { throw new NotImplementedException(); }
        public virtual IAbstractType InternalClone() { throw new NotImplementedException(); }

        public virtual void Write(IUpdateEncoder encoder) { throw new NotImplementedException(); }

        /// <summary>
        /// Call event listeners with an event. This will also add an event to all parents
        /// for observeDeep handlers.
        /// </summary>
        public virtual void CallTypeObservers(ITransaction transaction, IYEvent evt)
        {
            var type = this;

            while (true)
            {
                if (!transaction.ChangedParentTypes.TryGetValue(type, out var values))
                {
                    values = new List<IYEvent>();
                    transaction.ChangedParentTypes[type] = values;
                }

                values.Add(evt);

                if (type.Item == null)
                {
                    break;
                }

                type = type.Item.Parent as AbstractType;
            }

            InvokeEventHandlers(evt, transaction);
        }

        /// <summary>
        /// Creates YEvent and calls all type observers.
        /// Must be implemented by each type.
        /// </summary>
        public virtual void CallObserver(ITransaction transaction, ISet<string> parentSubs)
        {
            // Do nothing.
        }

        public IItem _First()
        {
            var n = Start;
            while (n != null && n.Deleted)
            {
                n = n.Right as IItem;
            }
            return n;
        }

        public void InvokeEventHandlers(IYEvent evt, ITransaction transaction)
        {
            EventHandler?.Invoke(this, new YEventArgs(evt, transaction));
        }

        public void CallDeepEventHandlerListeners(IList<IYEvent> events, ITransaction transaction)
        {
            DeepEventHandler?.Invoke(this, new YDeepEventArgs(events, transaction));
        }

        public string FindRootTypeKey()
        {
            return Doc.FindRootTypeKey(this);
        }

        protected void TypeMapDelete(ITransaction transaction, string key)
        {
            if (Map.TryGetValue(key, out var c))
            {
                c.Delete(transaction);
            }
        }

        protected void TypeMapSet(ITransaction transaction, string key, object value)
        {
            if (!Map.TryGetValue(key, out var left))
            {
                left = null;
            }

            var doc = transaction.Doc;
            var ownClientId = doc.ClientId;
            IContent content;

            if (value == null)
            {
                content = new ContentAny(new object[] { value });
            }
            else
            {
                switch (value)
                {
                    case YDoc d:
                        content = new ContentDoc(d);
                        break;
                    case AbstractType at:
                        content = new ContentType(at);
                        break;
                    case byte[] ba:
                        content = new ContentBinary(ba);
                        break;
                    default:
                        content = new ContentAny(new[] { value });
                        break;
                }
            }

            var newItem = new StructItem(new StructID(ownClientId, doc.Store.GetState(ownClientId)), left, left?.LastId, null, null, this, key, content);
            newItem.Integrate(transaction, 0);
        }

        protected bool TryTypeMapGet(string key, out object value)
        {
            if (Map.TryGetValue(key, out var val) && !val.Deleted)
            {
                value = val.Content.GetContent()[val.Length - 1];
                return true;
            }

            value = default;
            return false;
        }

        protected object TypeMapGetSnapshot(string key, ISnapshot snapshot)
        {
            if (!Map.TryGetValue(key, out var v))
            {
                v = null;
            }

            while (v != null && (!snapshot.StateVector.ContainsKey(v.Id.Client) || v.Id.Clock >= snapshot.StateVector[v.Id.Client]))
            {
                v = v.Left as IItem;
            }

            return v != null && v.IsVisible(snapshot) ? v.Content.GetContent()[v.Length - 1] : null;
        }

        protected IEnumerable<KeyValuePair<string, IItem>> TypeMapEnumerate() => Map.Where(kvp => !kvp.Value.Deleted);

        protected IEnumerable<KeyValuePair<string, object>> TypeMapEnumerateValues()
        {
            foreach (var kvp in TypeMapEnumerate())
            {
                var key = kvp.Key;
                var value = kvp.Value.Content.GetContent()[kvp.Value.Length - 1];
                yield return new KeyValuePair<string, object>(key, value);
            }
        }
    }
}
