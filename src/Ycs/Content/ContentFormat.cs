﻿﻿﻿// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;
using Ycs.Contracts;
using Ycs.Core;

namespace Ycs.Content
{
    public class ContentFormat : IContentEx
    {
        internal const int _ref = 6;

        public readonly string Key;
        public readonly object Value;

        internal ContentFormat(string key, object value)
        {
            Key = key;
            Value = value;
        }

        int IContentEx.Ref => _ref;

        public bool Countable => false;
        public int Length => 1;

        public IContent Copy() => new ContentFormat(Key, Value);

        public IReadOnlyList<object> GetContent() => throw new NotImplementedException();

        public IContent Splice(int offset) => throw new NotImplementedException();

        public bool MergeWith(IContent right)
        {
            return false;
        }

        void IContentEx.Integrate(ITransaction transaction, IStructItem item)
        {
            // Search markers are currently unsupported for rich text documents.
            // Check if parent implements array-like functionality and clear search markers if needed
            (item.Parent as IYArrayBase).ClearSearchMarkers();
        }

        void IContentEx.Delete(ITransaction transaction)
        {
            // Do nothing.
        }

        void IContentEx.Gc(IStructStore store)
        {
            // Do nothing.
        }

        void IContentEx.Write(IUpdateEncoder encoder, int offset)
        {
            encoder.WriteKey(Key);
            encoder.WriteJson(Value);
        }

        internal static ContentFormat Read(IUpdateDecoder decoder)
        {
            var key = decoder.ReadKey();
            var value = decoder.ReadJson();
            return new ContentFormat(key, value);
        }
    }
}
