// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System.Collections.Generic;

namespace Ycs.Contracts
{
    public class ChangesCollection
    {
        public ISet<IItem> Added;
        public ISet<IItem> Deleted;
        public IList<Delta> Delta;
        public IDictionary<string, ChangeKey> Keys;
    }

    public class Delta
    {
        public object Insert;
        public int? Delete;
        public int? Retain;
        public IDictionary<string, object> Attributes;
    }

    public enum ChangeAction
    {
        Add,
        Update,
        Delete
    }

    public class ChangeKey
    {
        public ChangeAction Action;
        public object OldValue;
    }
}
