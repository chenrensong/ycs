// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System.Collections.Generic;
using Ycs.Types;

namespace Ycs.Contracts
{
    public interface IYMap
    {
        int Count { get; }

        void CallObserver(ITransaction transaction, ISet<string> parentSubs);
        YMap Clone();
        bool ContainsKey(string key);
        void Delete(string key);
        object Get(string key);
        IEnumerator<KeyValuePair<string, object>> GetEnumerator();
        void Integrate(IYDoc doc, IItem item);
        IAbstractType InternalClone();
        IAbstractType InternalCopy();
        IEnumerable<string> Keys();
        void Set(string key, object value);
        IEnumerable<object> Values();
        void Write(IUpdateEncoder encoder);
    }
}