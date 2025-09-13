// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System.Collections.Generic;

namespace Ycs.Contracts
{
    public interface IAbstractType
    {
        IYDoc Doc { get; }
        IItem Item { get; set; }
        int Length { get; set; }
        IDictionary<string, IItem> Map { get; set; }
        IAbstractType Parent { get; }
        IItem Start { get; set; }
        void CallDeepEventHandlerListeners(IList<IYEvent> events, ITransaction transaction);
        void CallObserver(ITransaction transaction, ISet<string> parentSubs);
        void CallTypeObservers(ITransaction transaction, IYEvent evt);
        string FindRootTypeKey();
        void Integrate(IYDoc doc, IItem item);
        IAbstractType InternalClone();
        IAbstractType InternalCopy();
        void InvokeEventHandlers(IYEvent evt, ITransaction transaction);
        void Write(IUpdateEncoder encoder);
        IItem _First();
    }
}