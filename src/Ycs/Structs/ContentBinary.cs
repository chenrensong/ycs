// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;
using Ycs.Contracts;

namespace Ycs.Structs
{
    public class ContentBinary : IContentEx
    {
        internal const int _ref = 3;

        private readonly byte[] _content;

        internal ContentBinary(byte[] data)
        {
            _content = data;
        }

        int IContentEx.Ref => _ref;

        public bool Countable => true;
        public int Length => 1;

        public IReadOnlyList<object> GetContent() => new object[] { _content };

        public IContent Copy() => new ContentBinary(_content);

        public IContent Splice(int offset)
        {
            throw new NotImplementedException();
        }

        public bool MergeWith(IContent right)
        {
            return false;
        }

        void IContentEx.Integrate(ITransaction transaction, IItem item)
        {
            // Do nothing.
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
            encoder.WriteBuffer(_content);
        }

        internal static ContentBinary Read(IUpdateDecoder decoder)
        {
            var content = decoder.ReadBuffer();
            return new ContentBinary(content);
        }
    }
}
