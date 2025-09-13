﻿﻿﻿// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;
using Ycs.Contracts;

namespace Ycs.Content
{
    internal class ContentDoc : IContentEx
    {
        internal const int _ref = 9;

        internal ContentDoc(IYDoc doc)
        {
            if (doc.Item != null)
            {
                throw new Exception("This document was already integrated as a sub-document. You should create a second instance instead with the same guid.");
            }

            Doc = doc;
            Opts = new YDocOptions();

            if (!doc.Gc)
            {
                Opts.Gc = false;
            }

            if (doc.AutoLoad)
            {
                Opts.AutoLoad = true;
            }

            if (doc.Meta != null)
            {
                Opts.Meta = doc.Meta;
            }
        }

        int IContentEx.Ref => _ref;

        public bool Countable => true;
        public int Length => 1;

        public IYDoc Doc { get; internal set; }
        public YDocOptions Opts { get; internal set; } = new YDocOptions();

        public IReadOnlyList<object> GetContent() => new[] { Doc };

        public IContent Copy() => new ContentDoc(Doc);

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
            // This needs to be reflected in doc.destroy as well.
            Doc.Item = item;
            transaction.SubdocsAdded.Add(Doc);

            if (Doc.ShouldLoad)
            {
                transaction.SubdocsLoaded.Add(Doc);
            }
        }

        void IContentEx.Delete(ITransaction transaction)
        {
            if (transaction.SubdocsAdded.Contains(Doc))
            {
                transaction.SubdocsAdded.Remove(Doc);
            }
            else
            {
                transaction.SubdocsRemoved.Add(Doc);
            }
        }

        void IContentEx.Gc(IStructStore store)
        {
            // Do nothing.
        }

        void IContentEx.Write(IUpdateEncoder encoder, int offset)
        {
            // 32 digits separated by hyphens, no braces.
            encoder.WriteString(Doc.Guid);
            Opts.Write(encoder, offset);
        }

        private static Func<YDocOptions, IYDoc> _docFactory;
        
        internal static void SetDocFactory(Func<YDocOptions, IYDoc> factory)
        {
            _docFactory = factory;
        }
        
        internal static ContentDoc Read(IUpdateDecoder decoder)
        {
            var guidStr = decoder.ReadString();

            var opts = YDocOptions.Read(decoder);
            opts.Guid = guidStr;

            if (_docFactory == null)
            {
                throw new InvalidOperationException("DocFactory not initialized. Call YcsBootstrap.Initialize() first.");
            }
            
            return new ContentDoc(_docFactory(opts));
        }
    }
}
