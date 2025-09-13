// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;
using Ycs.Contracts;

namespace Ycs.Types
{
    public interface IAbstractType
    {
        IYDoc Doc { get; }
        IItem Item { get; set; }
        int Length { get; set; }
        IDictionary<string, IItem> Map { get; set; }
        IAbstractType Parent { get; }
        IItem Start { get; set; }
        void CallDeepEventHandlerListeners(IList<YEvent> events, ITransaction transaction);
        void CallObserver(ITransaction transaction, ISet<string> parentSubs);
        void CallTypeObservers(ITransaction transaction, YEvent evt);
        string FindRootTypeKey();
        void Integrate(IYDoc doc, IItem item);
        IAbstractType InternalClone();
        IAbstractType InternalCopy();
        void InvokeEventHandlers(YEvent evt, ITransaction transaction);
        void Write(IUpdateEncoder encoder);
        IItem _First();
    }
}