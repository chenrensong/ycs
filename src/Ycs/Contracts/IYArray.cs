// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System.Collections.Generic;

namespace Ycs.Contracts
{
    public interface IYArray : IAbstractType
    {
        int Length { get; }
        void Add(ICollection<object> content);
        void CallObserver(ITransaction transaction, ISet<string> parentSubs);
        IYArray Clone();
        void Delete(int index, int length = 1);
        object Get(int index);
        void Insert(int index, ICollection<object> content);
        void Integrate(IYDoc doc, IStructItem item);
        IAbstractType InternalClone();
        IAbstractType InternalCopy();
        IReadOnlyList<object> Slice(int start = 0);
        IReadOnlyList<object> Slice(int start, int end);
        IList<object> ToArray();
        void Unshift(ICollection<object> content);
        void Write(IUpdateEncoder encoder);
    }


    public interface IYArrayBase
    {
        void ClearSearchMarkers();
    }
}